package pull

import (
	"fmt"
	"tee/help"
	"tee/key"
	"tee/ocs"
	"tee/process/cache"
	pb "tee/proto"
	"tee/utils"

	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/protobuf/proto"
)

var cacheProgramInfo = make(map[common.Address]pb.Info)

func GetProgramInfo(programAddress common.Address) (*pb.Info, error) {
	// get from cache
	info := cache.GetProgramInfo(programAddress)
	if info != nil {
		// fmt.Println("Get program info from cache")
		return info, nil
	}

	contractAddr := common.HexToAddress(help.MCAddress)
	parsedABI := help.ParsedMCABI

	// get program info from contract
	var infoHashOut [32]byte
	err := help.CallContractMethod(parsedABI, contractAddr, "ProgramList", []interface{}{programAddress}, &infoHashOut)
	if err != nil {
		return nil, fmt.Errorf("failed to get program info: %v", err)
	}
	infoHash := infoHashOut[:]

	// get program info from off-chain
	encryptedInfo := ocs.GetInfo(programAddress, infoHash)
	if !key.MatchHash(encryptedInfo, infoHash) {
		return nil, fmt.Errorf("info hash mismatch")
	}

	// decypt result
	decryptedInfo, err := key.DecryptAES(encryptedInfo, key.KeyMgt)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt program information: %v", err)
	}

	// decode result pb to get program info
	var programInfo pb.Info
	err = proto.Unmarshal(decryptedInfo, &programInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to decode program information: %v", err)
	}

	// save to cache
	cacheProgramInfo[programAddress] = programInfo
	return &programInfo, nil
}

func GetProgramDetails(programAddress common.Address, stateKey string, codeKey string) ([]byte, []byte, error) {
	// get from cache
	code, states := cache.GetProgramDetails(programAddress)
	if code != nil && states != nil {
		// fmt.Println("Get program details from cache")
		return code, states, nil
	}

	// compatibal without statekey and codekey
	if stateKey == "" && codeKey == "" {
		info, err := GetProgramInfo(programAddress)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get program info: %v", err)
		}

		stateKey = info.Keys[len(info.Keys)-1]
		codeKey = info.CodeKey
	}

	parsedABI := help.ParsedMCABI
	MCAddress := common.HexToAddress(help.MCAddress)

	// get code hash from contract
	var codeHashOut [32]byte
	err := help.CallContractMethod(parsedABI, MCAddress, "ProgramCodes", []interface{}{programAddress}, &codeHashOut)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get code: %v", err)
	}
	codeHash := codeHashOut[:]

	// get code from off-chain
	encryptedCode := ocs.GetCode(programAddress)
	if !key.MatchHash(encryptedCode, codeHash) {
		return nil, nil, fmt.Errorf("code hash mismatch")
	}

	// decrypt code
	code, err = key.DecryptAES(encryptedCode, codeKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt code: %v", err)
	}

	// get states hash from contract
	var statesHashOut [32]byte
	err = help.CallContractMethod(parsedABI, MCAddress, "ProgramStates", []interface{}{programAddress}, &statesHashOut)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get states: %v", err)
	}
	statesHash := statesHashOut[:]

	// get states from off-chain
	encryptedStates := ocs.GetStates(programAddress, statesHash)
	if !key.MatchHash(encryptedStates, statesHash) {
		return nil, nil, fmt.Errorf("states hash mismatch")
	}

	// decrypt states
	states, err = key.DecryptAES(encryptedStates, stateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt states: %v", err)
	}

	return code, states, nil
}

func GetLatestExecutionBlock() (*utils.BlockInfo, error) {
	parsedABI := help.ParsedMCABI
	methodName := "latestExecutionBlock"

	// get latest execution block from contract
	var block utils.BlockInfo
	err := help.CallContractMethod(parsedABI, common.HexToAddress(help.MCAddress), methodName, nil, &block)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest execution block: %v", err)
	}

	return &block, nil
}
