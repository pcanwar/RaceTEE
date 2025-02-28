package operation

import (
	"context"
	"fmt"
	"log"

	"client/help"
	"client/key"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func Result(contractAddr common.Address) {
	client := help.Client
	parsedABI := help.ParsedClientABI

	// Construct a filter
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddr},
	}

	// Subscribe to logs
	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatalf("Failed to subscribe to logs: %v", err)
	}

	fmt.Println("Listening for Result events...")

	// Listening for Result events...
	for {
		select {
		case err := <-sub.Err():
			log.Fatalf("Subscription error: %v", err)
		case vLog := <-logs:
			// Check if the event is Result
			event := parsedABI.Events["Result"]
			if event.ID != vLog.Topics[0] {
				continue
			}

			// Unpack the log data
			var result struct {
				EncryptedResult    []byte
				EncryptedResultKey []byte
			}
			err := parsedABI.UnpackIntoInterface(&result, "Result", vLog.Data)
			if err != nil {
				log.Printf("Failed to unpack log data: %v", err)
				continue
			}

			// decrypt the result
			resultKey := key.GetResultKey(string(result.EncryptedResultKey))
			if resultKey != "" {
				decryptedResult, err := key.DecryptAES(result.EncryptedResult, resultKey)
				if err != nil {
					fmt.Printf("Failed to decrypt result: %v", err)
				}
				fmt.Printf("Result Event (%s): Result = %v\n", contractAddr.Hex(), decryptedResult)
			}
		}
	}
}
