package main

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"time"
)

type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
}

func (b *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	headers := append(timestamp, b.Data...)
	headers = append(headers, b.PrevBlockHash...)
	hash := sha256.Sum256(headers)
	b.Hash = hash[:]
}

type Blockchain struct {
	blocks []*Block
}

func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Data:          []byte(data),
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
	}
	block.SetHash()
	return block
}

func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}

func NewBlockchain() *Blockchain {
	return &Blockchain{[]*Block{NewGenesisBlock()}}
}

func (bc *Blockchain) AddBlock(data string) {
	prevBlock := bc.blocks[len(bc.blocks)-1]
	newBlock := NewBlock(data, prevBlock.Hash)
	bc.blocks = append(bc.blocks, newBlock)
}

func RunBlockchainOne() {
	fmt.Println("=== Simple Blockchain (Pattern 1) ===")
	bc := NewBlockchain()

	bc.AddBlock("Send 1 BTC to Alice")
	bc.AddBlock("Send 2 BTC to Bob")

	for _, block := range bc.blocks {
		fmt.Printf("Timestamp: %d\n", block.Timestamp)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("PrevBlockHash: %x\n", block.PrevBlockHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Println()
	}
}
