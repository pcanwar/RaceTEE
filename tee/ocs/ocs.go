// simulation for off-chain storage (should be implemented through disk rather than memory)
package ocs

import "github.com/ethereum/go-ethereum/common"

var codes = map[common.Address][]byte{}
var states = map[common.Address]map[string][]byte{}
var info = map[common.Address]map[string][]byte{}

func GetCode(addr common.Address) []byte {
	if code, exists := codes[addr]; exists {
		// return a copy of the code to prevent modification from outside
		return append([]byte{}, code...)
	}
	return nil
}

func SetCode(addr common.Address, code []byte) {
	if codes[addr] == nil {
		codes[addr] = code
	}
}

func GetStates(addr common.Address, hash []byte) []byte {
	if state, exists := states[addr][string(hash)]; exists {
		// return a copy of the state to prevent modification from outside
		return append([]byte{}, state...)
	}
	return nil
}

func SetStates(addr common.Address, hash []byte, state []byte) {
	if states[addr] == nil {
		states[addr] = map[string][]byte{}
	}
	states[addr][string(hash)] = state
}

func GetInfo(addr common.Address, hash []byte) []byte {
	if i, exists := info[addr][string(hash)]; exists {
		// return a copy of the list to prevent modification from outside
		return append([]byte{}, i...)
	}
	return nil
}

func SetInfo(addr common.Address, hash []byte, i []byte) {
	if info[addr] == nil {
		info[addr] = map[string][]byte{}
	}
	info[addr][string(hash)] = i
}
