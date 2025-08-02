package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"math"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/dgraph-io/badger/v3"
)

const targetBitsPersistent = 16
const dbFile = "./blockchain.db"

type BlockPersistent struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
	Difficulty    int
}

type BlockchainPersistent struct {
	lastHash []byte
	db       *badger.DB
}

type BlockchainIterator struct {
	currentHash []byte
	db          *badger.DB
}

type ProofOfWorkPersistent struct {
	block  *BlockPersistent
	target *big.Int
}

func (b *BlockPersistent) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

func DeserializeBlock(d []byte) *BlockPersistent {
	var block BlockPersistent

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}

func NewProofOfWorkPersistent(b *BlockPersistent) *ProofOfWorkPersistent {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBitsPersistent))

	pow := &ProofOfWorkPersistent{b, target}
	return pow
}

func (pow *ProofOfWorkPersistent) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.Data,
			[]byte(strconv.FormatInt(pow.block.Timestamp, 10)),
			[]byte(strconv.Itoa(targetBitsPersistent)),
			[]byte(strconv.Itoa(nonce)),
		},
		[]byte{},
	)
	return data
}

func (pow *ProofOfWorkPersistent) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Printf("Mining the block containing \"%s\"\n", pow.block.Data)
	for nonce < math.MaxInt64 {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			fmt.Printf("\r%x", hash)
			break
		} else {
			nonce++
		}
	}
	fmt.Print("\n\n")

	return nonce, hash[:]
}

func (pow *ProofOfWorkPersistent) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1
	return isValid
}

func NewBlockPersistent(data string, prevBlockHash []byte) *BlockPersistent {
	block := &BlockPersistent{
		Timestamp:     time.Now().Unix(),
		Data:          []byte(data),
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
		Nonce:         0,
		Difficulty:    targetBitsPersistent,
	}

	pow := NewProofOfWorkPersistent(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func NewGenesisBlockPersistent() *BlockPersistent {
	return NewBlockPersistent("Genesis Block", []byte{})
}

func NewBlockchainPersistent() *BlockchainPersistent {
	var lastHash []byte

	opts := badger.DefaultOptions(dbFile)
	opts.Logger = nil
	db, err := badger.Open(opts)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err == badger.ErrKeyNotFound {
			fmt.Println("No existing blockchain found. Creating a new one...")
			genesis := NewGenesisBlockPersistent()
			err := txn.Set(genesis.Hash, genesis.Serialize())
			if err != nil {
				log.Panic(err)
			}
			err = txn.Set([]byte("lh"), genesis.Hash)
			if err != nil {
				log.Panic(err)
			}
			lastHash = genesis.Hash
		} else {
			err := item.Value(func(val []byte) error {
				lastHash = append([]byte{}, val...)
				return nil
			})
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := &BlockchainPersistent{lastHash, db}
	return bc
}

func (bc *BlockchainPersistent) AddBlock(data string) {
	var lastHash []byte

	err := bc.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			log.Panic(err)
		}
		err = item.Value(func(val []byte) error {
			lastHash = append([]byte{}, val...)
			return nil
		})
		return err
	})

	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlockPersistent(data, lastHash)

	err = bc.db.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = txn.Set([]byte("lh"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}
		bc.lastHash = newBlock.Hash

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}

func (bc *BlockchainPersistent) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.lastHash, bc.db}
	return bci
}

func (i *BlockchainIterator) Next() *BlockPersistent {
	var block *BlockPersistent

	err := i.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(i.currentHash)
		if err != nil {
			log.Panic(err)
		}
		err = item.Value(func(val []byte) error {
			block = DeserializeBlock(val)
			return nil
		})
		return err
	})

	if err != nil {
		log.Panic(err)
	}

	i.currentHash = block.PrevBlockHash

	return block
}

func (bc *BlockchainPersistent) Close() {
	bc.db.Close()
}

func RunBlockchainThree() {
	fmt.Println("=== Blockchain with Persistent Storage (Pattern 3) ===")
	fmt.Println("Commands: add <data> | print")
	fmt.Println("Example: go run *.go 3 add \"Send 1 BTC to Alice\"")
	fmt.Println("Example: go run *.go 3 print")
	fmt.Println()

	bc := NewBlockchainPersistent()
	defer bc.Close()

	if len(os.Args) < 3 {
		fmt.Println("Usage:")
		fmt.Println("  go run *.go 3 add \"<data>\"")
		fmt.Println("  go run *.go 3 print")
		return
	}

	command := os.Args[2]

	switch command {
	case "add":
		if len(os.Args) < 4 {
			fmt.Println("Usage: go run *.go 3 add \"<data>\"")
			return
		}
		data := os.Args[3]
		bc.AddBlock(data)
		fmt.Printf("Block added: %s\n", data)

	case "print":
		bci := bc.Iterator()

		for {
			block := bci.Next()

			fmt.Printf("============ Block %x ============\n", block.Hash)
			fmt.Printf("Timestamp: %d\n", block.Timestamp)
			fmt.Printf("Data: %s\n", block.Data)
			fmt.Printf("PrevBlockHash: %x\n", block.PrevBlockHash)
			fmt.Printf("Hash: %x\n", block.Hash)
			fmt.Printf("Nonce: %d\n", block.Nonce)
			fmt.Printf("Difficulty: %d\n", block.Difficulty)

			pow := NewProofOfWorkPersistent(block)
			fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
			fmt.Println()

			if len(block.PrevBlockHash) == 0 {
				break
			}
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Available commands: add, print")
	}
}
