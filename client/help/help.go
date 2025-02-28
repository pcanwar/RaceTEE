package help

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"strings"

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

var (
	RPCURL        = "ws://127.0.0.1:8545"
	AccountsPath  = "./artifacts/accounts.json"
	MCAddressPath = "./artifacts/managementAddress.json"
	ClientABIPath = "./artifacts/ProgramContract.json"
	MCABIPath     = "./artifacts/ManagementContract.json"

	MCAddress string
	// PRGAddress      string
	Client          *ethclient.Client
	ParsedClientABI abi.ABI
	ParsedMCABI     abi.ABI
	Accounts        []Account
	ChainID         *big.Int
)

func init() {
	getClient()
	MCAddress = loadAddress(MCAddressPath)
	ParsedClientABI = LoadABI(ClientABIPath)
	ParsedMCABI = LoadABI(MCABIPath)
	Accounts = LoadAccounts()

	// ChainID = big.NewInt(1337) // for ganache
	ChainID = big.NewInt(31337) // for hardhat
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

// load ABI from JSON file
func LoadABI(path string) abi.ABI {
	// read file content
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

	// get ABI
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
func LoadBytecode(path string) []byte {
	// read file content
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

	return common.FromHex(bytecode.(string))
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

func LoadGolangCode(programmePath string) []byte {
	data, err := ioutil.ReadFile(programmePath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}
	return data
}
