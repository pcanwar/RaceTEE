package deploy

import (
	pb "client/proto"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"

	"client/help"
	"client/key"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"google.golang.org/protobuf/proto"
)

func DeployProgramme(code []byte, mainAccountIndex int, config pb.UserConfig) string {
	client := help.Client
	account := help.Accounts[mainAccountIndex]
	parsedABI := help.ParsedClientABI
	bytecode := help.LoadBytecode(help.ClientABIPath)

	// encrypt code and config
	encryptedCode, err := key.ECIESEncrypt(code)
	if err != nil {
		log.Fatalf("Failed to encrypt code: %v", err)
	}
	// encode config
	configBytes, err := proto.Marshal(&config)
	if err != nil {
		log.Fatalf("Failed to marshal config: %v", err)
	}
	encryptedConfig, err := key.ECIESEncrypt(configBytes)
	if err != nil {
		log.Fatalf("Failed to encrypt config: %v", err)
	}
	transactionKey := key.TXPubKeyBytes

	// encode the constructor arguments
	constructorArgs, err := parsedABI.Pack("", encryptedCode, encryptedConfig, transactionKey, common.HexToAddress(help.MCAddress))
	if err != nil {
		log.Fatalf("Failed to pack constructor arguments: %v", err)
	}

	// append bytecode and constructor arguments
	data := append(bytecode, constructorArgs...)

	// get nonce, gas price and gas limit
	nonce, err := client.PendingNonceAt(context.Background(), common.HexToAddress(account.Address))
	if err != nil {
		log.Fatalf("Failed to get nonce: %v", err)
	}
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf("Failed to get gas price: %v", err)
	}

	gasLimit := uint64(3000000)              // set enough gas limit
	value := big.NewInt(1000000000000000000) // in wei (1 eth)

	tx := types.NewContractCreation(nonce, value, gasLimit, gasPrice, data)

	// sign transaction
	privateKey := account.PrivateKey
	signedTx, err := help.SignTransaction(client, privateKey, tx)
	if err != nil {
		log.Fatalf("Failed to sign transaction: %v", err)
	}

	// send transaction
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatalf("Failed to send transaction: %v", err)
	}

	// get contract address
	contractAddress := crypto.CreateAddress(common.HexToAddress(account.Address), nonce)
	fmt.Printf("Contract deployed at address: %s\n", contractAddress.Hex())
	// save contract address to file
	// saveContractAddressToFile("./artifacts/programAddress.json", contractAddress)
	return contractAddress.Hex()
}

func saveContractAddressToFile(filename string, contractAddress common.Address) {
	output := map[string]string{
		"address": contractAddress.Hex(),
	}
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}
	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		log.Fatalf("Failed to write JSON file: %v", err)
	}
	fmt.Printf("Contract address saved to %s\n", filename)
}
