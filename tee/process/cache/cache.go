package cache

import (
	pb "tee/proto"

	"github.com/ethereum/go-ethereum/common"
)

type PRGCache struct {
	Code   []byte
	States []byte
}

// type pb.Info struct {
// 	Keys              []string `abi:"keys"`
// 	CodeKey           string   `abi:"codeKey"`
// 	HistoryKeyDiscard bool     `abi:"historyKeyDiscard"`
// 	KeyRotation       uint     `abi:"keyRotation"`
// 	ACL               []string `abi:"ACL"`
// 	ExecutionCount    uint     `abi:"executionCount"`
// 	Nounce            uint     `abi:"nounce"`
// }

// store the code and states of all program within one round of execution
var CacheStates = make(map[common.Address]PRGCache)
var CacheInfos = make(map[common.Address]*pb.Info)

func GetProgramDetails(programAddress common.Address) ([]byte, []byte) {
	if cache, ok := CacheStates[programAddress]; ok {
		return cache.Code, cache.States
	}
	return nil, nil
}

func GetProgramInfo(programAddress common.Address) *pb.Info {
	if cache, ok := CacheInfos[programAddress]; ok {
		return cache
	}
	return nil
}

func SetBatchProgramDetails(addrs []common.Address, codes [][]byte, allStates [][]byte) {
	for i, addr := range addrs {
		SetProgramDetails(addr, codes[i], allStates[i])
	}
}

func SetProgramDetails(programAddress common.Address, code []byte, states []byte) {
	CacheStates[programAddress] = PRGCache{Code: code, States: states}
}

func SetProgramInfo(programAddress common.Address, info *pb.Info) {
	CacheInfos[programAddress] = info
}

func ClearCache() {
	CacheStates = make(map[common.Address]PRGCache)
	CacheInfos = make(map[common.Address]*pb.Info)
}
