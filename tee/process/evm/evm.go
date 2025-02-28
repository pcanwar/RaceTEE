package evm

import (
	"fmt"
	"math/big"
	"tee/help"
	"tee/pull"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
)

const getStatesFunc = "getStates"
const setStatesFunc = "setStates"
const getInteractContractsFunc = "getInteractContracts"
const gas = 90000000000 // set a large gas limit

var contractAddress common.Address
var callerAddress common.Address
var vmConfig = vm.Config{}
var chainConfig = params.MainnetChainConfig
var evmContext = vm.BlockContext{
	CanTransfer: func(db vm.StateDB, from common.Address, amount *uint256.Int) bool {
		return db.GetBalance(from).Cmp(amount) >= 0
	},
	Transfer: func(db vm.StateDB, from common.Address, to common.Address, amount *uint256.Int) {
		db.SubBalance(from, amount, tracing.BalanceChangeUnspecified)
		db.AddBalance(to, amount, tracing.BalanceChangeUnspecified)
	},
	GetHash: nil,

	Coinbase:    common.Address{},
	GasLimit:    uint64(0),
	BlockNumber: big.NewInt(0),
	Time:        uint64(0),
	Difficulty:  big.NewInt(0),
	BaseFee:     big.NewInt(0), // No base fee
	BlobBaseFee: big.NewInt(0), // No blob base fee
	// Random:      &common.Hash{},

}

// initialize EVM environment
var statedb *state.StateDB
var evm *vm.EVM

var txContext = vm.TxContext{
	Origin:   callerAddress,
	GasPrice: big.NewInt(0),
}

func SetConfig(_blockNumber *big.Int, _blockTime uint64, _contractAddress common.Address, _callerAddress common.Address) {
	contractAddress = _contractAddress
	callerAddress = _callerAddress
	evmContext.BlockNumber = _blockNumber
	evmContext.Time = _blockTime
	txContext.Origin = _callerAddress
}

func refresh() {
	var err error
	statedb, err = state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
	if err != nil {
		panic(err)
	}
	evm = vm.NewEVM(evmContext, txContext, statedb, chainConfig, vmConfig)
}

func Deploy(userCode []byte) ([]byte, []byte, error) {
	refresh()
	// deploy code
	code, address, _, err := evm.Create(vm.AccountRef(callerAddress), userCode, uint64(gas), uint256.MustFromBig(big.NewInt(0)))
	if err != nil {
		fmt.Println("Error create contract:", err)
		return nil, nil, err
	}

	newStates, err := getStates(address)
	if err != nil {
		fmt.Println("Error getting states:", err)
		return nil, nil, err
	}
	return newStates, code, nil
}

func Execute(userCode []byte, states []byte, input []byte) ([]common.Address, [][]byte, [][]byte, interface{}, error) {
	refresh()
	// load interact contracts
	contracts, codes, err := loadInteractContracts(contractAddress)
	if err != nil {
		fmt.Println("Error loading interact contracts:", err)
		return nil, nil, nil, nil, err
	}

	// execute contract in inner EVM
	result, _, err := evm.Call(vm.AccountRef(callerAddress), contractAddress, input, uint64(gas), uint256.MustFromBig(big.NewInt(0)))
	if err != nil {
		fmt.Println("Error executing contract:", err)
		return nil, nil, nil, nil, err
	}

	// get all states
	newAllStates, err := getAllStates(contracts)
	if err != nil {
		fmt.Println("Error getting states:", err)
		return nil, nil, nil, nil, err
	}

	return contracts, newAllStates, codes, result, nil
}

func loadInteractContracts(contractAddress common.Address) ([]common.Address, [][]byte, error) {
	// getcontract Details
	code, states, err := pull.GetProgramDetails(contractAddress, "", "")
	if err != nil {
		fmt.Println("Error getting contract details:", err)
		return nil, nil, err
	}
	// deploy contract
	statedb.SetCode(contractAddress, code)
	err = setStates(contractAddress, states)
	if err != nil {
		fmt.Println("Error setting interactContract states:", err)
		return nil, nil, err
	}
	// store all interactive contracts address and code
	contracts := []common.Address{contractAddress}
	codes := [][]byte{code}

	// get interact contracts
	getInteractContractsInput, err := help.ParsedSystemABI.Pack(getInteractContractsFunc)
	if err != nil {
		fmt.Printf("Failed to pack getInteractContracts function call: %v", err)
	}
	result, _, err := evm.Call(vm.AccountRef(callerAddress), contractAddress, getInteractContractsInput, uint64(gas), uint256.MustFromBig(big.NewInt(0)))
	if err != nil {
		fmt.Println("Error executing getInteractContracts function call:", err)
		return nil, nil, err
	}

	// decode result
	res, err := help.ParsedSystemABI.Unpack(getInteractContractsFunc, result)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unpack: %v", err)
	}

	if len(res) <= 0 {
		return nil, nil, fmt.Errorf("empty unpacked result")
	}

	// get first return value and ensure its type is []common.Address
	addrs, ok := res[0].([]common.Address) // 确保类型匹配
	if !ok {
		return nil, nil, fmt.Errorf("unexpected type: %T", res[0])
	}
	for _, addr := range addrs {
		subAddrs, subCodes, err := loadInteractContracts(addr)
		if err != nil {
			fmt.Println("Error loading interact contracts:", err)
			return nil, nil, err
		}
		contracts = append(contracts, subAddrs...)
		codes = append(codes, subCodes...)
	}

	return contracts, codes, nil
}

func getAllStates(contractAddr []common.Address) ([][]byte, error) {
	var allStates [][]byte
	for _, addr := range contractAddr {
		states, err := getStates(addr)
		if err != nil {
			fmt.Println("Error getting states:", err)
			return nil, err
		}
		allStates = append(allStates, states)
	}
	return allStates, nil
}

func getStates(contractAddr common.Address) ([]byte, error) {
	// get current states
	getStatesInput, err := help.ParsedSystemABI.Pack(getStatesFunc)
	if err != nil {
		fmt.Printf("Failed to pack function call: %v", err)
	}
	// execute contract in inner EVM
	result, _, err := evm.Call(vm.AccountRef(callerAddress), contractAddr, getStatesInput, uint64(gas), uint256.MustFromBig(big.NewInt(0)))
	if err != nil {
		fmt.Println("Error executing contract:", err)
		return nil, err
	}

	// unpack result
	res, err := help.ParsedSystemABI.Unpack(getStatesFunc, result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack: %v", err)
	}

	// get first return value and ensure its type is []byte
	if len(res) > 0 {
		newStates, ok := res[0].([]byte)
		if !ok {
			return nil, fmt.Errorf("unexpected type: %T", res[0])
		}
		return newStates, nil
	}
	return nil, fmt.Errorf("empty unpacked result")
}

func setStates(contractAddr common.Address, states []byte) error {
	// set new states
	setStatesInput, err := help.ParsedSystemABI.Pack(setStatesFunc, states)
	if err != nil {
		fmt.Printf("Failed to pack function call: %v", err)
		return err
	}
	_, _, err = evm.Call(vm.AccountRef(callerAddress), contractAddr, setStatesInput, uint64(gas), uint256.MustFromBig(big.NewInt(0)))
	if err != nil {
		fmt.Println("Error executing contract (setStates):", err)
		return err
	}

	return nil
}
