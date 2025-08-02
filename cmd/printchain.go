package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var printchainCmd = &cobra.Command{
	Use:   "printchain",
	Short: "Print all blocks in the blockchain",
	Long: `Print all blocks in the blockchain from genesis block to the latest block.
This command reads data from the BadgerDB database and displays block information.`,
	Run: func(cmd *cobra.Command, args []string) {
		printChain()
	},
}

func printChain() {
	bc := NewBlockchainCLI()
	defer bc.Close()

	bci := bc.Iterator()

	// Store blocks to reverse order (genesis first)
	var blocks []*BlockCLI

	for {
		block := bci.Next()
		blocks = append(blocks, block)

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	// Print blocks from genesis to latest
	for i := len(blocks) - 1; i >= 0; i-- {
		block := blocks[i]

		fmt.Printf("============ Block %x ============\n", block.Hash)
		fmt.Printf("Timestamp: %d\n", block.Timestamp)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("PrevBlockHash: %x\n", block.PrevBlockHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Nonce: %d\n", block.Nonce)
		fmt.Printf("Difficulty: %d\n", block.Difficulty)

		pow := NewProofOfWorkCLI(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
	}
}

func init() {
	rootCmd.AddCommand(printchainCmd)
}
