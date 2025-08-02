package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run *.go <pattern>")
		fmt.Println("Available patterns:")
		fmt.Println("  1 - Simple Blockchain")
		fmt.Println("  2 - Blockchain with Proof of Work")
		fmt.Println("  3 - Blockchain with Persistent Storage")
		fmt.Println("  4 - Blockchain with CLI Interface")
		fmt.Println("  5 - Blockchain with Wallet System")
		fmt.Println("  6 - Blockchain with Transactions and UTXO")
		fmt.Println("  7 - Blockchain with P2P Network Layer")
		return
	}

	pattern := os.Args[1]

	switch pattern {
	case "1":
		RunBlockchainOne()
	case "2":
		RunBlockchainTwo()
	case "3":
		RunBlockchainThree()
	case "4":
		RunBlockchainFour()
	case "5":
		RunBlockchainFive()
	case "6":
		RunBlockchainSix()
	case "7":
		runBlockchainSeven()
	default:
		fmt.Printf("Unknown pattern: %s\n", pattern)
		fmt.Println("Available patterns: 1, 2, 3, 4, 5, 6, 7")
	}
}
