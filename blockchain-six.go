package main

import (
	"fmt"
	"os"

	"blockchain-app/wallet"
)

func RunBlockchainSix() {
	fmt.Println("=== Blockchain with Transactions and UTXO (Pattern 6) ===")
	fmt.Println("This pattern demonstrates the foundation for transaction functionality.")
	fmt.Println("Key Features Implemented:")
	fmt.Println("  ✓ Transaction structure (TxInput, TxOutput, Transaction)")
	fmt.Println("  ✓ UTXO (Unspent Transaction Output) model")
	fmt.Println("  ✓ Digital signature framework (ECDSA)")
	fmt.Println("  ✓ Coinbase transactions (initial coin creation)")
	fmt.Println("  ✓ Transaction validation and verification")
	fmt.Println("  ✓ Wallet integration")
	fmt.Println()
	fmt.Println("Available wallet commands:")
	fmt.Println("  go run *.go 6 createwallet")
	fmt.Println("  go run *.go 6 listaddresses")
	fmt.Println()

	if len(os.Args) < 3 {
		return
	}

	command := os.Args[2]

	switch command {
	case "createwallet":
		createWalletTX()
	case "listaddresses":
		listAddressesTX()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Available commands: createwallet, listaddresses")
	}
}


func createWalletTX() {
	wallets, _ := wallet.NewWallets()
	address := wallets.CreateWallet()
	wallets.SaveToFile()

	fmt.Printf("Your new address: %s\n", address)
}

func listAddressesTX() {
	wallets, _ := wallet.NewWallets()
	addresses := wallets.GetAddresses()

	if len(addresses) == 0 {
		fmt.Println("No wallet addresses found. Create a wallet first.")
		return
	}

	fmt.Println("Wallet addresses:")
	for _, address := range addresses {
		fmt.Printf("  %s\n", address)
	}
}