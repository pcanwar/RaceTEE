package help

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"

	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Account struct {
	Address    string `json:"address"`
	PrivateKey string `json:"privateKey"`
}

type Address struct {
	Address string `json:"address"`
}

type Output struct {
	ProgramAddress     common.Address `abi:"programAddress"`
	Info               [32]byte       `abi:"info"`
	States             [32]byte       `abi:"states"`
	Result             []byte         `abi:"result"`
	EncryptedResultKey []byte         `abi:"encryptedResultKey"`
	Code               [32]byte       `abi:"code"`
	TransType          uint8          `abi:"transType"`
}

var (
	Lang string

	RPCURL         = "ws://127.0.0.1:8545"
	MCABIPath      = "./artifacts/ManagementContract.json"
	ProgramABIPath = "./artifacts/StandardProgramContract.json"
	SystemABIPath  = "./artifacts/SystemContract.json"
	AccountsPath   = "./artifacts/accounts.json"
	MCAddressPath  = "./artifacts/managementAddress.json"

	MCAddress        string
	Client           *ethclient.Client
	ParsedMCABI      abi.ABI
	ParsedProgramABI abi.ABI
	ParsedSystemABI  abi.ABI
	Accounts         []Account

	TransTypeDeploy    uint8
	TransTypeExecution uint8
	TransTypeInteract  uint8
	TransTypeACL       uint8
	TransTypeError     uint8

	ChainID      *big.Int
	AccountIndex int

	AverageTimes int
)

func init() {
	TransTypeExecution = uint8(0)
	TransTypeDeploy = uint8(1)
	TransTypeInteract = uint8(2)
	TransTypeACL = uint8(3)
	TransTypeError = uint8(4)
	getClient()
	MCAddress = loadAddress(MCAddressPath)
	ParsedMCABI = LoadABI(MCABIPath)
	ParsedProgramABI = LoadABI(ProgramABIPath)
	ParsedSystemABI = LoadABI(SystemABIPath)
	Accounts = LoadAccounts()

	// ChainID = big.NewInt(1337) // for ganache
	ChainID = big.NewInt(31337)
}

func getClient() *ethclient.Client {
	client, err := ethclient.Dial(RPCURL)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	Client = client
	return client
}

// load accounts from JSON file
func LoadAccounts() []Account {
	data, err := ioutil.ReadFile(AccountsPath)
	if err != nil {
		log.Fatalf("Failed to read JSON file: %v", err)
	}

	// parse JSON
	var accounts []Account
	err = json.Unmarshal(data, &accounts)
	if err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}
	return accounts
}

func LoadABI(path string) abi.ABI {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read JSON file: %v", err)
	}

	// parse JSON
	var contractConfig map[string]interface{}
	err = json.Unmarshal(data, &contractConfig)
	if err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	// get ABI field
	abiField, ok := contractConfig["abi"]
	if !ok {
		log.Fatalf("ABI field not found in JSON")
	}

	// marshal ABI field
	abiJSON, err := json.Marshal(abiField)
	if err != nil {
		log.Fatalf("Failed to marshal ABI field: %v", err)
	}

	// parse ABI
	parsedABI, err := abi.JSON(strings.NewReader(string(abiJSON)))
	if err != nil {
		log.Fatalf("Failed to parse ABI: %v", err)
	}
	return parsedABI
}

// signs a transaction with the given private key
func SignTransaction(client *ethclient.Client, privateKeyHex string, tx *types.Transaction) (*types.Transaction, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %v", err)
	}

	var chainID *big.Int
	if ChainID != nil {
		chainID = ChainID
	} else {
		chainID, err = client.NetworkID(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to get chain ID: %v", err)
		}
	}

	return types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
}

// load bytcode from JSON file
func LoadBytecode(path string) string {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read JSON file: %v", err)
	}

	// parse JSON
	var contractConfig map[string]interface{}
	err = json.Unmarshal(data, &contractConfig)
	if err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	// get bytecode
	bytecode, ok := contractConfig["bytecode"]
	if !ok {
		log.Fatalf("bytecode field not found in JSON")
	}

	return bytecode.(string)
}

func loadAddress(filePath string) string {
	// read and parse JSON file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read JSON file: %v", err)
	}
	var result Address
	err = json.Unmarshal(data, &result)
	if err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	return result.Address
}

func ErrorOutput(errMsg string, programAddress common.Address, key []byte) Output {
	output := Output{
		ProgramAddress:     programAddress,
		TransType:          TransTypeError,
		Result:             []byte(errMsg),
		EncryptedResultKey: key,
	}
	return output
}

func ByteToByte32(b []byte) [32]byte {
	var bytes32 [32]byte
	copy(bytes32[:], b)
	return bytes32
}

func CallContractMethod(parsedABI abi.ABI, contractAddr common.Address, methodName string, params []interface{}, output interface{}) error {
	// encode call data
	callData, err := parsedABI.Pack(methodName, params...)
	if err != nil {
		return fmt.Errorf("failed to pack %s call data: %v", methodName, err)
	}

	// call contract
	result, err := Client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &contractAddr,
		Data: callData,
	}, nil)
	if err != nil {
		return fmt.Errorf("failed to call %s: %v", methodName, err)
	}

	// decode result
	err = parsedABI.UnpackIntoInterface(output, methodName, result)
	if err != nil {
		return fmt.Errorf("failed to unpack %s result: %v", methodName, err)
	}

	return nil
}
