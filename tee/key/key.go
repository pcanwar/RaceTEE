package key

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
)

var (
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
)

var mapTXKey = make(map[string]*ecies.PrivateKey)

var KeyMgt string

func init() {
	var err error
	PrivateKey, err = generateECDHKey()
	PublicKey = &PrivateKey.PublicKey
	if err != nil {
		panic(err)
	}
	KeyMgt, err = loadKeyFromFile("./key/tempMgtKey.json")
	if err != nil {
		panic(err)
	}
	_, _, err = loadTXKeyFromFile("./key/tempTxKey.json")
	if err != nil {
		panic(err)
	}
}

// GenerateECDHKey generates an ECDH private key
func generateECDHKey() (*ecdsa.PrivateKey, error) {
	privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ECDH private key: %v", err)
	}
	return privateKey, nil
}

func FormatECDSAPublicKey(pubKey *ecdsa.PublicKey) []byte {
	// format: X(32 bytes) || Y(32 bytes)
	return append(pubKey.X.Bytes(), pubKey.Y.Bytes()...)
}

func PublicKeyToAddress(pubKey *ecdsa.PublicKey) common.Address {
	pubBytes := FormatECDSAPublicKey(pubKey)
	hash := crypto.Keccak256(pubBytes)
	return common.BytesToAddress(hash[12:]) // take the last 20 bytes
}

func TEESign(message []byte) ([]byte, error) {
	signature, err := crypto.Sign(message, PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign message: %v", err)
	}
	return signature, nil
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

// AES encrypt
func EncryptAES(plainText []byte, key string) ([]byte, error) {
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

	// padding
	plainText = pkcs7Padding(plainText, block.BlockSize())

	// create cipher text
	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// create encrypt stream (CFB mode)
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainText)

	// return cipher text
	return cipherText, nil
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

// PKCS7 padding
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// PKCS7 unpadding
func pkcs7UnPadding(data []byte) []byte {
	length := len(data)
	unpadding := int(data[length-1])
	return data[:(length - unpadding)]
}

func encodeKey(key []byte) string {
	return base64.StdEncoding.EncodeToString(key)
}

func decodeKey(encodedKey string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encodedKey)
}

// KeyFile represents the structure for storing the AES key in a JSON file.
type KeyFile struct {
	Key string `json:"key"`
}

// LoadKeyFromFile reads the AES key from a JSON file.
func loadKeyFromFile(filePath string) (string, error) {
	// Read the JSON file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read key file: %w", err)
	}

	// Unmarshal the JSON into a KeyFile structure
	var keyFile KeyFile
	err = json.Unmarshal(data, &keyFile)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal key from JSON: %w", err)
	}

	return keyFile.Key, nil
}

// KeyFile represents the structure for storing the AES key in a JSON file.
type TXKeyFile struct {
	Private string `json:"private"`
	Public  string `json:"public"`
}

// read the AES key (fixed key, not implement the TXKey change for prototype) and parse it into an ECIES key pair
func loadTXKeyFromFile(filePath string) (*ecies.PrivateKey, *ecies.PublicKey, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read key file: %w", err)
	}
	var txKeyFile TXKeyFile
	err = json.Unmarshal(data, &txKeyFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal key from JSON: %w", err)
	}

	// parse to appropriate format
	privKeyBytes, err := hex.DecodeString(txKeyFile.Private)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode private key hex: %w", err)
	}
	pubKeyBytes, err := hex.DecodeString(txKeyFile.Public)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode public key hex: %w", err)
	}

	// convert to ECDSA private key
	privateKey, err := crypto.ToECDSA(privKeyBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert to ECDSA private key: %w", err)
	}

	// convert to ECDSA public key
	publicKey, err := crypto.UnmarshalPubkey(pubKeyBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert to ECDSA public key: %w", err)
	}

	// convert to ECIES key pair
	eciesPrivateKey := ecies.ImportECDSA(privateKey)
	eciesPublicKey := ecies.ImportECDSAPublic((*ecdsa.PublicKey)(publicKey))

	mapTXKey[hex.EncodeToString(pubKeyBytes)] = eciesPrivateKey

	return eciesPrivateKey, eciesPublicKey, nil
}

func ECIESDecrypt(cipherText []byte, pubKey string) ([]byte, error) {
	TXPrivateKey := mapTXKey[pubKey]
	if TXPrivateKey == nil {
		return nil, fmt.Errorf("failed to get TXPrivateKey")
	}
	return TXPrivateKey.Decrypt(cipherText, nil, nil)
}

func GetHash(data []byte) []byte {
	hash := crypto.Keccak256(data)
	return hash
}

func MatchHash(data []byte, hash []byte) bool {
	return bytes.Equal(GetHash(data), hash)
}
