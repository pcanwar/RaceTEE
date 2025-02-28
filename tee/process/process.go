package process

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"tee/help"
	"tee/key"
	"tee/process/cache"
	"tee/utils"
)

//	type Input struct {
//		FuncName string          `json:"funcName"`
//		Args     json.RawMessage `json:"args"`
//	}
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

func Process(events []map[string]interface{}) []help.Output {
	outputs := []help.Output{}
	for _, event := range events {
		eventName := event["eventName"].(string)
		var _outputs []help.Output

		switch eventName {
		case "Deploy":
			println("Deploy")
			_outputs = Deploy(event)
		case "Execution":
			println("Execution")
			_outputs = Execute(event)
		default:
			continue
		}
		outputs = append(outputs, _outputs...)
	}

	// clear cache
	cache.ClearCache()
	return outputs
}

var lastNonce uint64 = 0

func SendOutputsToChain(account help.Account, outputs []help.Output, startBlock, endBlock uint64) error {
	// Create a shared context
	ctx := context.Background()
	client, parsedABI := help.Client, help.ParsedMCABI

	// Get start and end block data
	start, err := utils.GetBlock(startBlock)
	if err != nil {
		return fmt.Errorf("failed to get start block: %v", err)
	}
	end, err := utils.GetBlock(endBlock)
	if err != nil {
		return fmt.Errorf("failed to get end block: %v", err)
	}

	// Generate hash of outputs and sign it
	hashOutputs, err := getHashOutputs(start, end, outputs)
	if err != nil {
		return fmt.Errorf("failed to hash outputs: %v", err)
	}
	signature, err := key.TEESign(hashOutputs)
	if err != nil {
		return fmt.Errorf("failed to sign outputs: %v", err)
	}

	// Encode transaction data with the contract ABI
	outputsEncoded, err := parsedABI.Pack("output", start, end, outputs, signature)
	if err != nil {
		return fmt.Errorf("failed to encode outputs: %v", err)
	}

	// Get nonce and ensure it increases sequentially
	if nonce == 0 {
		initNonce(common.HexToAddress(account.Address))
	}

	// Suggest gas price and add 1 Gwei
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return fmt.Errorf("failed to suggest gas price: %v", err)
	}
	gasPrice = new(big.Int).Add(gasPrice, big.NewInt(1000000000))

	// Estimate gas limit and add a 5% buffer
	MCAddress := common.HexToAddress(help.MCAddress)
	msg := ethereum.CallMsg{
		From:  common.HexToAddress(account.Address),
		To:    &MCAddress,
		Value: big.NewInt(0),
		Data:  outputsEncoded,
	}
	gasLimit, err := client.EstimateGas(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to estimate gas: %v", err)
	}
	gasLimit += gasLimit / 20

	// Create and sign the transaction
	tx := types.NewTransaction(nonce, MCAddress, big.NewInt(0), gasLimit, gasPrice, outputsEncoded)
	nonce++
	signedTx, err := help.SignTransaction(client, account.PrivateKey, tx)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %v", err)
	}

	// Send the transaction
	if err = client.SendTransaction(ctx, signedTx); err != nil {
		return fmt.Errorf("failed to send transaction: %v", err)
	}

	fmt.Printf("Transaction sent! Tx hash: %s\n", signedTx.Hash().Hex())
	return nil
}

// generate hash of the outputs by calling on-chain contract function for following signature
func getHashOutputs(startBlock utils.BlockInfo, endBlock utils.BlockInfo, outputs []help.Output) ([]byte, error) {
	parsedABI := help.ParsedMCABI

	// get hash of outputs
	var hashOutputs [32]byte
	err := help.CallContractMethod(parsedABI, common.HexToAddress(help.MCAddress), "hashOutputs", []interface{}{startBlock, endBlock, outputs}, &hashOutputs)
	if err != nil {
		return nil, fmt.Errorf("failed to get hash of outputs: %v", err)
	}

	return hashOutputs[:], nil
}
