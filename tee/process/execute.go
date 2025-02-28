package process

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"tee/help"
	"tee/key"
	"tee/ocs"
	"tee/process/cache"
	"tee/process/compacity"
	"tee/pull"
	"tee/utils"

	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/exp/rand"
	"google.golang.org/protobuf/proto"
)

func Execute(event map[string]interface{}) []help.Output {
	data := event["data"].(map[string]interface{})
	encryptedResultKey := data["encryptedResultKey"].([]byte)
	programAddress := data["programAddress"].(common.Address)

	// get program info, prepare for execution
	info, err := pull.GetProgramInfo(programAddress)
	if err != nil {
		fmt.Printf("Failed to get program info: %v", err)
		return []help.Output{help.ErrorOutput("Failed to get program info", programAddress, encryptedResultKey)}
	}

	stateKey := info.Keys[len(info.Keys)-1]
	codeKey := info.CodeKey
	// get program details
	code, states, err := pull.GetProgramDetails(programAddress, stateKey, codeKey)
	if err != nil {
		fmt.Printf("Failed to get program details: %v", err)
		return []help.Output{help.ErrorOutput("Failed to get program details", programAddress, encryptedResultKey)}
	}

	// get result key
	txPubKey := data["transactionKey"].([]byte)
	txPubKeyStr := hex.EncodeToString(txPubKey)
	resultKey, err := key.ECIESDecrypt(encryptedResultKey, txPubKeyStr)
	if err != nil {
		fmt.Printf("Failed to decrypt result key: %v", err)
		return []help.Output{help.ErrorOutput("Failed to decrypt result key", programAddress, encryptedResultKey)}
	}

	// parse input
	encryptedinput := data["encryptedInput"].([]byte)
	input, err := key.ECIESDecrypt(encryptedinput, txPubKeyStr)
	if err != nil {
		fmt.Printf("Failed to decrypt input: %v", err)
		return []help.Output{help.ErrorOutput("Failed to decrypt input", programAddress, encryptedResultKey)}
	}

	// execute the program
	conf := compacity.GetCompacityConfig(event)
	addresses, newStates, codes, result, err := compacity.Execute(code, states, input, conf)
	if err != nil {
		fmt.Printf("Failed to execute program: %v", err)
		return []help.Output{help.ErrorOutput("Failed to execute program", programAddress, encryptedResultKey)}
	}

	// save new states to cache
	cache.SetBatchProgramDetails(addresses, codes, newStates)

	// prepare output
	caller := data["caller"].(common.Address).String()
	outputs := prepareOutput(addresses, newStates, result, resultKey, programAddress, encryptedResultKey, caller)
	return outputs
}

// Function to prepare output
func prepareOutput(addresses []common.Address, newStates [][]byte, result interface{}, resultKey []byte, programAddress common.Address, encryptedResultKey []byte, caller string) []help.Output {
	res, err := toBytes(result)
	if err != nil {
		fmt.Printf("Failed to convert result: %v", err)
		return []help.Output{help.ErrorOutput("Failed to convert result", programAddress, encryptedResultKey)}
	}
	// encrypt result
	encryptedResult, err := key.EncryptAES(res, string(resultKey))
	if err != nil {
		fmt.Printf("Failed to encrypt result: %v", err)
		return []help.Output{help.ErrorOutput("Failed to encrypt result", programAddress, encryptedResultKey)}
	}

	var outputs []help.Output
	for i, addr := range addresses {
		state := newStates[i]

		info, err := pull.GetProgramInfo(addr)
		if err != nil {
			fmt.Printf("Failed to get program info: %v", err)
			return []help.Output{help.ErrorOutput("Failed to get program info", programAddress, encryptedResultKey)}
		}

		// check if caller is in ACL
		ALC := info.ACL
		if len(ALC) != 0 && !utils.Contains(ALC, caller) {
			fmt.Printf("Caller %v is not in ACL", caller)
			return []help.Output{help.ErrorOutput("Caller is not in ACL", programAddress, encryptedResultKey)}
		}

		info.Nounce = uint32(rand.Intn(1000000)) // set nounce to a random number
		info.ExecutionCount += 1                 // increase executionCount
		// rotate key
		if info.KeyRotation != 0 && info.ExecutionCount%info.KeyRotation == 0 {
			k, error := key.GenerateAESKey()
			if error != nil {
				panic(fmt.Sprintf("Failed to generate AES key: %v", error))
			}
			if info.HistoryKeyDiscard {
				info.Keys = []string{string(k)}
			} else {
				keys := info.Keys
				keys = append(keys, string(k))
				info.Keys = keys
			}
		}
		infoBytes, err := proto.Marshal(info) // encode info
		if err != nil {
			panic(fmt.Sprintf("Failed to encode info: %v", err))
		}

		// prepare output
		encryptedInfo, err := key.EncryptAES(infoBytes, key.KeyMgt)
		if err != nil {
			panic(fmt.Sprintf("Failed to encrypt info: %v", err))
		}

		stateKey := info.Keys[len(info.Keys)-1]
		encryptedStates, err := key.EncryptAES([]byte(state), stateKey)
		if err != nil {
			panic(fmt.Sprintf("Failed to encrypt states: %v", err))
		}

		// save states off-chain
		statesHash := key.GetHash(encryptedStates)
		ocs.SetStates(addr, statesHash, encryptedStates)

		// save info off-chain
		infoHash := key.GetHash(encryptedInfo)
		ocs.SetInfo(addr, infoHash, encryptedInfo)

		// prepare output
		var output help.Output
		if i == 0 {
			output = help.Output{
				TransType:          help.TransTypeExecution,
				ProgramAddress:     addr,
				Info:               help.ByteToByte32(infoHash),
				States:             help.ByteToByte32(statesHash),
				Result:             encryptedResult,
				EncryptedResultKey: encryptedResultKey,
			}
		} else {
			output = help.Output{
				TransType:      help.TransTypeInteract,
				ProgramAddress: addr,
				Info:           help.ByteToByte32(infoHash),
				States:         help.ByteToByte32(statesHash),
				// Code:               [32]byte{}, // do not need to return code
				// Result:             []byte{}, // do not need to return result
				// EncryptedResultKey: []byte{}, // do not need to return result key
			}
		}
		outputs = append(outputs, output)

		// save info to cache
		cache.SetProgramInfo(addr, info)
	}
	return outputs
}

// general interface{} to []byte
func toBytes(input interface{}) ([]byte, error) {
	if input == nil {
		return nil, nil
	}

	if b, ok := input.([]byte); ok {
		return b, nil
	}

	if str, ok := input.(string); ok {
		return []byte(str), nil
	}

	switch v := input.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		return []byte(fmt.Sprintf("%v", v)), nil
	}

	if jsonBytes, err := json.Marshal(input); err == nil {
		return jsonBytes, nil
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(input); err != nil {
		return nil, fmt.Errorf("failed to encode input: %v", err)
	}
	return buf.Bytes(), nil
}
