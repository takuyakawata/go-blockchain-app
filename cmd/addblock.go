package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var data string

var addblockCmd = &cobra.Command{
	Use:   "addblock",
	Short: "Add a new block to the blockchain",
	Long: `Add a new block to the blockchain with the specified data.
The data will be stored in the block and the block will be mined using Proof of Work.`,
	Run: func(cmd *cobra.Command, args []string) {
		addBlock(data)
	},
}

func addBlock(data string) {
	if data == "" {
		fmt.Println("Error: Data cannot be empty")
		fmt.Println("Usage: blockchain addblock --data \"Your data here\"")
		return
	}

	bc := NewBlockchainCLI()
	defer bc.Close()

	bc.AddBlock(data)
	fmt.Println("Success! Block added to the blockchain.")
}

func init() {
	rootCmd.AddCommand(addblockCmd)
	addblockCmd.Flags().StringVarP(&data, "data", "d", "", "Block data (required)")
	addblockCmd.MarkFlagRequired("data")
}
