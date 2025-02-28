package operation

import (
	"context"
	"fmt"
	"math/big"
	"tee/help"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func CallRegister(attestationReport, key []byte, depositAmount *big.Int, account help.Account) error {
	client := help.Client
	parsedABI := help.ParsedMCABI

	// prepare register call data
	callData, err := parsedABI.Pack("register", attestationReport, key)
	if err != nil {
		return fmt.Errorf("failed to pack register call data: %v", err)
	}

	// create transaction
	nonce, err := client.PendingNonceAt(context.Background(), common.HexToAddress(account.Address))
	if err != nil {
		return fmt.Errorf("failed to get nonce: %v", err)
	}
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get gas price: %v", err)
	}
	tx := types.NewTransaction(nonce, common.HexToAddress(help.MCAddress), depositAmount, 5000000, gasPrice, callData)

	// sign transaction
	signedTx, err := help.SignTransaction(client, account.PrivateKey, tx)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %v", err)
	}

	// send transaction
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %v", err)
	}

	fmt.Printf("Register transaction sent! Tx hash: %s\n", signedTx.Hash().Hex())
	return nil
}

func CallWithdraw(signature []byte, account help.Account) error {
	client := help.Client
	parsedABI := help.ParsedMCABI

	// create withdraw call data
	callData, err := parsedABI.Pack("withdraw", signature)
	if err != nil {
		return fmt.Errorf("failed to pack withdraw call data: %v", err)
	}

	// create transaction
	nonce, err := client.PendingNonceAt(context.Background(), common.HexToAddress(account.Address))
	if err != nil {
		return fmt.Errorf("failed to get nonce: %v", err)
	}
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get gas price: %v", err)
	}
	tx := types.NewTransaction(nonce, common.HexToAddress(help.MCAddress), big.NewInt(0), 3000000, gasPrice, callData)

	// sign transaction
	signedTx, err := help.SignTransaction(client, account.PrivateKey, tx)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %v", err)
	}

	// send transaction
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %v", err)
	}

	fmt.Printf("Withdraw transaction sent! Tx hash: %s\n", signedTx.Hash().Hex())
	return nil
}
