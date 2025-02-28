package process

import (
	"encoding/hex"
	"fmt"
	"tee/help"
	"tee/key"
	"tee/ocs"
	"tee/process/cache"
	"tee/process/compacity"
	pb "tee/proto"

	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/exp/rand"
	"google.golang.org/protobuf/proto"
)

func Deploy(event map[string]interface{}) []help.Output {
	data := event["data"].(map[string]interface{})

	// deploy program
	conf := compacity.GetCompacityConfig(event)
	encryptedCode := data["encryptedCode"].([]byte)
	pubKey := data["transactionKey"].([]byte)
	code, err := key.ECIESDecrypt(encryptedCode, hex.EncodeToString(pubKey))

	if err != nil {
		fmt.Printf("Failed to execute decrypt code: %v", err)
		return []help.Output{}
	}

	states, newCode, err := compacity.Deploy(code, conf)
	if err != nil {
		fmt.Printf("Failed to deploy program: %v", err)
		return []help.Output{}
	}

	// get user config
	encryptedConfig := data["encryptedConfig"].([]byte)
	configBytes, err := key.ECIESDecrypt(encryptedConfig, hex.EncodeToString(pubKey))
	if err != nil {
		fmt.Printf("Failed to decrypt config: %v", err)
		return []help.Output{}
	}
	var userConfig pb.UserConfig
	err = proto.Unmarshal(configBytes, &userConfig)
	if err != nil {
		fmt.Printf("Failed to unmarshal userConfig: %v", err)
		return []help.Output{}
	}

	// set info field
	stateKey, err := key.GenerateAESKey()
	if err != nil {
		panic(fmt.Sprintf("Failed to generate state key: %v", err))
	}
	codeKey, err := key.GenerateAESKey()
	if err != nil {
		panic(fmt.Sprintf("Failed to generate code key: %v", err))
	}
	info := &pb.Info{
		Keys:              []string{stateKey},
		CodeKey:           codeKey,
		HistoryKeyDiscard: userConfig.HistoryKeyDiscard,
		KeyRotation:       userConfig.KeyRotation,
		ACL:               userConfig.ACL,
		ExecutionCount:    0,
		// random Nounce for each prevent leakages
		Nounce: uint32(rand.Intn(1000000)),
	}
	infoBytes, err := proto.Marshal(info)
	if err != nil {
		panic(fmt.Sprintf("Failed to encode info: %v", err))
	}

	// prepare output
	encryptedInfo, err := key.EncryptAES(infoBytes, key.KeyMgt)
	if err != nil {
		panic(fmt.Sprintf("Failed to encrypt info: %v", err))
	}
	encryptedStates, err := key.EncryptAES([]byte(states), stateKey)
	if err != nil {
		panic(fmt.Sprintf("Failed to encrypt states: %v", err))
	}
	newEncryptedCode, err := key.EncryptAES(newCode, codeKey)
	if err != nil {
		panic(fmt.Sprintf("Failed to encrypt code: %v", err))
	}
	programAddress := data["programAddress"].(common.Address)
	// store code and states off-chain
	codeHash := key.GetHash(newEncryptedCode)
	statesHash := key.GetHash(encryptedStates)
	infoHash := key.GetHash(encryptedInfo)
	ocs.SetCode(programAddress, newEncryptedCode)
	ocs.SetStates(programAddress, statesHash, encryptedStates)
	ocs.SetInfo(programAddress, infoHash, encryptedInfo)
	// prepare output
	output := help.Output{
		TransType:      help.TransTypeDeploy,
		ProgramAddress: programAddress,
		Info:           help.ByteToByte32(infoHash),
		Code:           help.ByteToByte32(codeHash),
		States:         help.ByteToByte32(statesHash),
		// Result:             []byte{},
		// EncryptedResultKey: []byte{},
	}

	// save new states to cache
	cache.SetProgramDetails(programAddress, newCode, states)
	// save info to cache
	cache.SetProgramInfo(programAddress, info)
	return []help.Output{output}
}
