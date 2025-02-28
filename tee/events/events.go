package events

// event structure
// {
// 	"eventName": "Deploy",
// 	"data": {
// 		"encryptedCode": "0x1234",
// 		"encryptedConfig": "0x5678",
// 		"transactionKey": "0x9abc",
// 	    "programAddress": "0xdef0",
//      "caller": "0x1234",
// 	},
// 	"blockNumber": 12345,
// 	"txHash": "0x1234",
// 	"logIndex": 0
// }

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"tee/help"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func GetEventsFrom(start uint64, end uint64) []map[string]interface{} {
	client := help.Client
	parsedMCABI := help.ParsedMCABI

	logs, err := getEvents(client, start, end)
	if err != nil {
		log.Fatalf("Failed to get logs: %v", err)
	}
	events := parseLogs(client, parsedMCABI, logs)

	return events
}

func getEvents(client *ethclient.Client, fromBlock uint64, toBlock uint64) ([]types.Log, error) {
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(toBlock)),
		Addresses: []common.Address{common.HexToAddress(help.MCAddress)},
	}

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to filter logs: %v", err)
	}
	return logs, nil
}

func parseLogs(client *ethclient.Client, parsedABI abi.ABI, logs []types.Log) []map[string]interface{} {
	var parsedEvents []map[string]interface{}

	for _, vLog := range logs {
		eventName := ""
		switch vLog.Topics[0].Hex() {
		case parsedABI.Events["Deploy"].ID.Hex():
			eventName = "Deploy"
		case parsedABI.Events["Execution"].ID.Hex():
			eventName = "Execution"
		default:
			continue // ignore unknown event
		}

		// parse log data
		data := map[string]interface{}{
			"eventName": eventName,
			"data":      map[string]interface{}{},
		}
		err := parsedABI.UnpackIntoMap(data["data"].(map[string]interface{}), eventName, vLog.Data)
		if err != nil {
			log.Printf("Failed to unpack log data for %s: %v", eventName, err)
			continue
		}

		blockNumber := big.NewInt(int64(vLog.BlockNumber))
		block, err := client.BlockByNumber(context.Background(), blockNumber)
		if err != nil {
			log.Printf("Failed to get block by number: %v", err)
			continue
		}

		// add addional information
		data["blockNumber"] = blockNumber
		data["blockTime"] = block.Time()
		data["blockHash"] = vLog.BlockHash.Hex()
		data["txHash"] = vLog.TxHash.Hex()
		data["logIndex"] = vLog.Index

		parsedEvents = append(parsedEvents, data)
	}

	return parsedEvents
}
