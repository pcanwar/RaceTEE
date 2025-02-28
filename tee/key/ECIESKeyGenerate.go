// Important: This file is used to generate ECIES key pair (private key and public key) and put them in the tempTxKey.json file.
// It will not be used during the execution of the prototype.
//
// Generate ECIES key pair (private key and public key) and put them in the tempTxKey.json file.
// public key is copied and used in ManagementContract for user to encrypt the transaction
package key

// package main

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
)

func main() {
	// generate ECDSA key pair
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal("Failed to generate private key:", err)
	}
	publicKey := &privateKey.PublicKey

	// convert ECDSA key pair to ECIES key pair
	eciesPrivateKey := ecies.ImportECDSA(privateKey)
	eciesPublicKey := ecies.ImportECDSAPublic(publicKey)

	// output the keys
	fmt.Println("Private Key (Hex):", hex.EncodeToString(crypto.FromECDSA(privateKey)))
	fmt.Println("Public Key (Hex):", hex.EncodeToString(crypto.FromECDSAPub(publicKey)))

	// verify the keys conversion
	fmt.Println("ECIES Private Key:", eciesPrivateKey)
	fmt.Println("ECIES Public Key:", eciesPublicKey)
}
