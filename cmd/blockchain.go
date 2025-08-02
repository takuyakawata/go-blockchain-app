package cmd

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"math"
	"math/big"
	"strconv"
	"time"

	"github.com/dgraph-io/badger/v3"
)

const targetBitsCLI = 16
const dbFileCLI = "./blockchain-cli.db"

type BlockCLI struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
	Difficulty    int
}

type BlockchainCLI struct {
	lastHash []byte
	db       *badger.DB
}

type BlockchainIteratorCLI struct {
	currentHash []byte
	db          *badger.DB
}

type ProofOfWorkCLI struct {
	block  *BlockCLI
	target *big.Int
}

func (b *BlockCLI) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

func DeserializeBlockCLI(d []byte) *BlockCLI {
	var block BlockCLI

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}

func NewProofOfWorkCLI(b *BlockCLI) *ProofOfWorkCLI {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBitsCLI))

	pow := &ProofOfWorkCLI{b, target}
	return pow
}

func (pow *ProofOfWorkCLI) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.Data,
			[]byte(strconv.FormatInt(pow.block.Timestamp, 10)),
			[]byte(strconv.Itoa(targetBitsCLI)),
			[]byte(strconv.Itoa(nonce)),
		},
		[]byte{},
	)
	return data
}

func (pow *ProofOfWorkCLI) Run() (int, []byte) {
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

func (pow *ProofOfWorkCLI) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1
	return isValid
}

func NewBlockCLI(data string, prevBlockHash []byte) *BlockCLI {
	block := &BlockCLI{
		Timestamp:     time.Now().Unix(),
		Data:          []byte(data),
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
		Nonce:         0,
		Difficulty:    targetBitsCLI,
	}

	pow := NewProofOfWorkCLI(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func NewGenesisBlockCLI() *BlockCLI {
	return NewBlockCLI("Genesis Block", []byte{})
}

func NewBlockchainCLI() *BlockchainCLI {
	var lastHash []byte

	opts := badger.DefaultOptions(dbFileCLI)
	opts.Logger = nil
	db, err := badger.Open(opts)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err == badger.ErrKeyNotFound {
			fmt.Println("No existing blockchain found. Creating a new one...")
			genesis := NewGenesisBlockCLI()
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

	bc := &BlockchainCLI{lastHash, db}
	return bc
}

func (bc *BlockchainCLI) AddBlock(data string) {
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

	newBlock := NewBlockCLI(data, lastHash)

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

func (bc *BlockchainCLI) Iterator() *BlockchainIteratorCLI {
	bci := &BlockchainIteratorCLI{bc.lastHash, bc.db}
	return bci
}

func (i *BlockchainIteratorCLI) Next() *BlockCLI {
	var block *BlockCLI

	err := i.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(i.currentHash)
		if err != nil {
			log.Panic(err)
		}
		err = item.Value(func(val []byte) error {
			block = DeserializeBlockCLI(val)
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

func (bc *BlockchainCLI) Close() {
	bc.db.Close()
}
