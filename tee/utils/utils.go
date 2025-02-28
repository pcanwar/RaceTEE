package utils

import (
	"context"
	"fmt"
	"math/big"
	"tee/help"
)

type BlockInfo struct {
	BlockNumber uint64   `abi:"blockNumber"`
	BlockHash   [32]byte `abi:"blockHash"`
}

// Get the block information by block number
func GetBlock(num uint64) (BlockInfo, error) {
	client := help.Client

	blockNum := big.NewInt(0).SetUint64(num)
	block, err := client.BlockByNumber(context.Background(), blockNum)
	if err != nil {
		return BlockInfo{}, fmt.Errorf("failed to fetch block: %v", err)
	}

	var blockHash [32]byte
	copy(blockHash[:], block.Hash().Bytes())
	res := BlockInfo{
		BlockNumber: num,
		BlockHash:   blockHash,
	}
	return res, nil
}

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
