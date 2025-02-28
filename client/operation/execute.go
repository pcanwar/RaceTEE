package operation

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"client/help"
	"client/key"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var nonce uint64 = 0

func initNonce(address common.Address) {
	client := help.Client
	ctx := context.Background()
	_nonce, err := client.PendingNonceAt(ctx, address)
	if err != nil {
		log.Fatalf("Failed to get nonce: %v", err)
	}
	nonce = _nonce
}

func Execute(contractAddress common.Address, accountNum int, input []byte) {
	parsedABI := help.ParsedClientABI
	// encode the execution call
	resultKey, err := key.GenerateAESKey()
	if err != nil {
		log.Fatalf("Failed to generate result key: %v", err)
	}
	encryptedResultKey, err := key.ECIESEncrypt([]byte(resultKey))
	if err != nil {
		log.Fatalf("Failed to encrypt result key: %v", err)
	}
	transactionKey := key.TXPubKeyBytes
	encryptedInput, err := key.ECIESEncrypt(input)
	if err != nil {
		log.Fatalf("Failed to encrypt input: %v", err)
	}
	data, err := parsedABI.Pack("execution", encryptedInput, encryptedResultKey, transactionKey)
	if err != nil {
		log.Fatalf("Failed to pack execution call data: %v", err)
	}

	BaseExeuction(contractAddress, accountNum, data)

	// cache the result key
	key.SaveResultKey(string(encryptedResultKey), resultKey)
}

func BaseExeuction(contractAddress common.Address, accountNum int, data []byte) {
	client := help.Client
	ctx := context.Background()
	account := help.Accounts[accountNum]

	value := big.NewInt(0) // eth value to send

	// create a transaction
	// contractAddress := common.HexToAddress(help.PRGAddress)
	if nonce == 0 {
		initNonce(common.HexToAddress(account.Address))
	}

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		log.Fatalf("Failed to get gas price: %v", err)
	}

	msg := ethereum.CallMsg{
		From:  common.HexToAddress(account.Address),
		To:    &contractAddress,
		Value: value,
		Data:  data,
	}
	gasLimit, err := client.EstimateGas(ctx, msg)
	if err != nil {
		log.Fatalf("Failed to estimate gas: %v", err)
	}
	gasLimit = gasLimit + gasLimit/20 // add 5% buffer
	tx := types.NewTransaction(nonce, contractAddress, value, gasLimit, gasPrice, data)
	nonce++

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

	fmt.Printf("Transaction sent! Tx hash: %s\n", signedTx.Hash().Hex())
}
