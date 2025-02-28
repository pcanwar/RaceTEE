package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"strconv"
	"tee/events"
	"tee/help"
	"tee/key"
	"tee/operation"
	"tee/process"
	"tee/pull"

	"github.com/ethereum/go-ethereum/core/types"
)

// ./tee -lang s
func main() {
	var lang string
	var i string
	flag.StringVar(&lang, "lang", "s", "User program language: g(golang) or s(solidity)")
	flag.StringVar(&i, "i", "5", "Account index")
	flag.Parse()
	help.Lang = lang
	help.AccountIndex, _ = strconv.Atoi(i)
	register()
	// wait for the TEE to be registered
	// time.Sleep(20 * time.Second)
	start()
}

func register() {
	account := help.Accounts[help.AccountIndex]

	// Register the TEE on chain
	teePK := key.FormatECDSAPublicKey(key.PublicKey)
	// teePkHash := sha256.Sum256(teePK)
	// localQuote := quote.GetQuote(teePkHash[:])
	// generate a fake quote
	localQuote := make([]byte, 0)
	_, err := rand.Read(localQuote)
	if err != nil {
		panic(err)
	}
	err = operation.CallRegister(localQuote, teePK, big.NewInt(1000000000000000000), account)
	if err != nil {
		panic(err)
	}
}

func start() {
	client := help.Client
	account := help.Accounts[help.AccountIndex]

	// Get the latest block number to run TEE first time
	blockNumber, err := client.BlockNumber(context.Background())
	if err != nil {
		log.Fatalf("Failed to get the latest block number: %v", err)
	}
	running(account, blockNumber)

	// Create a channel to receive new block headers
	headers := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Fatalf("Failed to subscribe to new head: %v", err)
	}
	defer sub.Unsubscribe()

	// Process each new block as it arrives
	for {
		select {
		case err := <-sub.Err():
			log.Fatalf("Error while subscribing to new head: %v", err)
		case header := <-headers:
			running(account, header.Number.Uint64())
		}
	}
}

func running(account help.Account, end uint64) {
	startBlock, err := pull.GetLatestExecutionBlock()
	if err != nil {
		panic(err)
	}
	// retrieve all events from the last execution block to the current block
	latest := (*startBlock).BlockNumber
	start := latest + 1
	fmt.Printf("Start: %v, End: %v\n", start, end)
	eventsList := events.GetEventsFrom(start, end)
	fmt.Printf("evetnsLength: %v\n", len(eventsList))
	// if there are no events, return
	if len(eventsList) == 0 {
		return
	}

	// process all events
	outputs := process.Process(eventsList)
	err = process.SendOutputsToChain(account, outputs, latest, end)
	if err != nil {
		panic(err)
	}
}
