package key

import (
	"client/help"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
)

var TXPubKey *ecies.PublicKey
var TXPubKeyBytes []byte
var cacheResultKey map[string]string

func init() {
	cacheResultKey = make(map[string]string)
	var err error
	TXPubKey, TXPubKeyBytes, err = GetTXPubKey()
	if err != nil {
		panic("Failed to get TXPubKey")
	}
}

func GetTXPubKey() (*ecies.PublicKey, []byte, error) {
	// get on-chain transaction key
	client := help.Client
	parsedABI := help.ParsedMCABI
	// encode the execution call
	callData, err := parsedABI.Pack("transactionPubKey")
	if err != nil {
		log.Fatalf("Failed to pack execution call data: %v", err)
	}
	// call contract
	MCAddress := common.HexToAddress(help.MCAddress)
	result, err := client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &MCAddress,
		Data: callData,
	}, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to call ProgramList: %v", err)
	}
	// decode result
	var pubkey string
	err = parsedABI.UnpackIntoInterface(&pubkey, "transactionPubKey", result)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unpack transactionPubKey result: %v", err)
	}

	// string to ecies publickey
	pubKeyBytes, err := hex.DecodeString(pubkey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode public key hex: %w", err)
	}

	// decode to ECDSA public key
	publicKey, err := crypto.UnmarshalPubkey(pubKeyBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert to ECDSA public key: %w", err)
	}

	// change to ECIES public key
	eciesPublicKey := ecies.ImportECDSAPublic((*ecdsa.PublicKey)(publicKey))

	return eciesPublicKey, pubKeyBytes, nil
}

func ECIESEncrypt(data []byte) ([]byte, error) {
	return ecies.Encrypt(rand.Reader, TXPubKey, data, nil, nil)
}

// generates a random AES key
func GenerateAESKey() (string, error) {
	// Validate key size
	keySize := 32
	// Create a slice to hold the key
	key := make([]byte, keySize)

	// Fill the key slice with cryptographically secure random bytes
	_, err := rand.Read(key)
	if err != nil {
		return "", fmt.Errorf("failed to generate random key: %w", err)
	}

	return encodeKey(key), nil
}

// AES decrypt
func DecryptAES(cipherText []byte, key string) ([]byte, error) {
	// decode key
	decodedKey, err := decodeKey(key)
	if err != nil {
		return nil, err
	}
	// create AES block cipher
	block, err := aes.NewCipher(decodedKey)
	if err != nil {
		return nil, err
	}

	// check cipher text length
	if len(cipherText) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// get IV
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	// create stream
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	// unpadding
	data := pkcs7UnPadding(cipherText)

	return data, nil
}

func encodeKey(key []byte) string {
	return base64.StdEncoding.EncodeToString(key)
}

func decodeKey(encodedKey string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encodedKey)
}

// PKCS7 unpadding
func pkcs7UnPadding(data []byte) []byte {
	length := len(data)
	unpadding := int(data[length-1])
	return data[:(length - unpadding)]
}

func SaveResultKey(encryptedKey string, key string) {
	cacheResultKey[encryptedKey] = key
}

func GetResultKey(encryptedKey string) string {
	key := cacheResultKey[encryptedKey]
	if key != "" {
		delete(cacheResultKey, encryptedKey)
	}
	return key
}
