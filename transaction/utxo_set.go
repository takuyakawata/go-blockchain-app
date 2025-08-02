package transaction

import (
	"encoding/hex"
	"log"

	"github.com/dgraph-io/badger/v3"
)

const utxoBucket = "chainstate"

// UTXOSet represents UTXO set
type UTXOSet struct {
	Blockchain *Blockchain
}

// TXOutputs collects TXOutput
type TXOutputs struct {
	Outputs []TXOutput
}

// Serialize serializes TXOutputs
func (outs TXOutputs) Serialize() []byte {
	return []byte{} // Simplified for now
}

// DeserializeOutputs deserializes TXOutputs
func DeserializeOutputs(data []byte) TXOutputs {
	return TXOutputs{} // Simplified for now
}

// FindSpendableOutputs finds and returns unspent outputs to reference in inputs
func (u UTXOSet) FindSpendableOutputs(pubkeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	db := u.Blockchain.db

	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(utxoBucket)
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()
			err := item.Value(func(v []byte) error {
				outs := DeserializeOutputs(v)

				txID := hex.EncodeToString(key[len(utxoBucket):])

				for outIdx, out := range outs.Outputs {
					if out.IsLockedWithKey(pubkeyHash) && accumulated < amount {
						accumulated += out.Value
						unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return accumulated, unspentOutputs
}

// FindUTXO finds UTXO for a public key hash
func (u UTXOSet) FindUTXO(pubKeyHash []byte) []TXOutput {
	var UTXOs []TXOutput
	db := u.Blockchain.db

	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(utxoBucket)
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(v []byte) error {
				outs := DeserializeOutputs(v)

				for _, out := range outs.Outputs {
					if out.IsLockedWithKey(pubKeyHash) {
						UTXOs = append(UTXOs, out)
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return UTXOs
}

// CountTransactions returns the number of transactions in the UTXO set
func (u UTXOSet) CountTransactions() int {
	db := u.Blockchain.db
	counter := 0

	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(utxoBucket)
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			counter++
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return counter
}

// Reindex rebuilds the UTXO set
func (u UTXOSet) Reindex() {
	db := u.Blockchain.db
	bucketName := []byte(utxoBucket)

	err := db.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = bucketName
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			err := txn.Delete(it.Item().Key())
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	UTXO := u.Blockchain.FindUTXO()

	err = db.Update(func(txn *badger.Txn) error {
		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID)
			if err != nil {
				return err
			}
			key = append(bucketName, key...)

			err = txn.Set(key, outs.Serialize())
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}

// Update updates the UTXO set with transactions from the Block
func (u UTXOSet) Update(block *Block) {
	db := u.Blockchain.db

	err := db.Update(func(txn *badger.Txn) error {
		for _, tx := range block.Transactions {
			if tx.IsCoinbase() == false {
				for _, vin := range tx.Vin {
					updatedOuts := TXOutputs{}
					outsBytes, err := txn.Get(append([]byte(utxoBucket), vin.Txid...))
					if err != nil {
						continue
					}
					err = outsBytes.Value(func(v []byte) error {
						outs := DeserializeOutputs(v)

						for outIdx, out := range outs.Outputs {
							if outIdx != vin.Vout {
								updatedOuts.Outputs = append(updatedOuts.Outputs, out)
							}
						}
						return nil
					})
					if err != nil {
						return err
					}

					if len(updatedOuts.Outputs) == 0 {
						err := txn.Delete(append([]byte(utxoBucket), vin.Txid...))
						if err != nil {
							return err
						}
					} else {
						err := txn.Set(append([]byte(utxoBucket), vin.Txid...), updatedOuts.Serialize())
						if err != nil {
							return err
						}
					}
				}
			}

			newOutputs := TXOutputs{}
			for _, out := range tx.Vout {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}

			err := txn.Set(append([]byte(utxoBucket), tx.ID...), newOutputs.Serialize())
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}
