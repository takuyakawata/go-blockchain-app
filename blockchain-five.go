package main

import (
	"blockchain-app/wallet"
	"fmt"
	"log"
	"os"
)

func RunBlockchainFive() {
	fmt.Println("=== Blockchain with Wallet System (Pattern 5) ===")
	fmt.Println("This pattern provides wallet functionality with ECDSA key pairs.")
	fmt.Println("To use wallet commands:")
	fmt.Println("  go run *.go 5 createwallet")
	fmt.Println("  go run *.go 5 listaddresses")
	fmt.Println()

	if len(os.Args) < 3 {
		fmt.Println("Usage:")
		fmt.Println("  go run *.go 5 createwallet")
		fmt.Println("  go run *.go 5 listaddresses")
		return
	}

	command := os.Args[2]

	switch command {
	case "createwallet":
		createWallet()
	case "listaddresses":
		listAddresses()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Available commands: createwallet, listaddresses")
	}
}

func createWallet() {
	wallets, _ := wallet.NewWallets()
	address := wallets.CreateWallet()
	wallets.SaveToFile()

	fmt.Printf("Your new address: %s\n", address)
}

func listAddresses() {
	wallets, err := wallet.NewWallets()
	if err != nil {
		log.Panic(err)
	}
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
