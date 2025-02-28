package compacity

import (
	"math/big"
	"tee/help"
	"tee/process/evm"
	"tee/process/golang"
	pb "tee/proto"

	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/protobuf/proto"
)

type Config struct {
	ProgramAddress common.Address
	Caller         common.Address
	BlockNumber    *big.Int
	BlockTime      uint64
}

var VMGolang = "golang"
var VMSolidity = "solidity"

func vm() string {
	if help.Lang == "g" {
		return VMGolang
	}
	return VMSolidity
}

func Deploy(code []byte, conf Config) ([]byte, []byte, error) {
	vm := vm()
	var states []byte
	var newCode []byte
	var err error
	switch vm {
	case VMGolang:
		states, newCode, err = deployGolang(code)
	case VMSolidity:
		states, newCode, err = deploySolidity(code, conf)
	}
	return states, newCode, err
}

func Execute(code []byte, states []byte, input []byte, conf Config) ([]common.Address, [][]byte, [][]byte, interface{}, error) {
	vm := vm()
	var addresses []common.Address
	var newStates [][]byte
	var codes [][]byte
	var result interface{}
	var err error
	switch vm {
	case VMGolang:
		addresses, newStates, codes, result, err = executeGolang(code, states, input, conf)
	case VMSolidity:
		addresses, newStates, codes, result, err = executeSolidity(code, states, input, conf)
	}
	return addresses, newStates, codes, result, err
}

func deploySolidity(code []byte, conf Config) ([]byte, []byte, error) {
	evm.SetConfig(conf.BlockNumber, conf.BlockTime, conf.ProgramAddress, conf.Caller)
	states, newCode, err := evm.Deploy(code)
	return states, newCode, err
}

func deployGolang(code []byte) ([]byte, []byte, error) {
	states, err := golang.Deploy(code)
	return states, code, err
}

func executeGolang(code []byte, states []byte, input []byte, conf Config) ([]common.Address, [][]byte, [][]byte, interface{}, error) {
	// parse input
	var decodedInput pb.GolangInput
	err := proto.Unmarshal(input, &decodedInput)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	// execute the program
	newStates, result, err := golang.Execute(code, states, decodedInput.FuncName, string(decodedInput.Args))
	addresses := []common.Address{conf.ProgramAddress}
	resStates := [][]byte{newStates}
	codes := [][]byte{code}
	return addresses, resStates, codes, result, err
}

func executeSolidity(code []byte, states []byte, input []byte, conf Config) ([]common.Address, [][]byte, [][]byte, interface{}, error) {
	evm.SetConfig(conf.BlockNumber, conf.BlockTime, conf.ProgramAddress, conf.Caller)
	// newStates, result, err := evm.Execute(code, states, input)
	return evm.Execute(code, states, input)
}

func GetCompacityConfig(event map[string]interface{}) Config {
	data := event["data"].(map[string]interface{})
	// get program info, prepare for deploy
	programAddress := data["programAddress"].(common.Address)
	callerAddress := data["caller"].(common.Address)
	blockNumber := event["blockNumber"].(*big.Int)
	// set block number
	// less than 12965000, set block number to 12965000, since lower block number has some unexpected behavior
	if blockNumber.Cmp(big.NewInt(12965000)) < 0 {
		blockNumber = big.NewInt(12965000)
	}
	blockTime := event["blockTime"].(uint64)
	conf := Config{
		ProgramAddress: programAddress,
		Caller:         callerAddress,
		BlockNumber:    blockNumber,
		BlockTime:      blockTime,
	}
	return conf
}
