package blockchain

import (
	"encoding/hex"
	"github.com/boltdb/bolt"
	"log"
)

type UTXOSet struct {
	Blockchain *BlockChain
}

const utxoBucket = "chainstate"

func (u UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	db := u.Blockchain.Db
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucket))
		c := bucket.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k)
			outs := DeserializeOutputs(v)

			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
					accumulated += out.Value
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Panicln(err)
	}
	return accumulated, unspentOutputs
}

func (u UTXOSet) FindUTXO(pubKeyHash []byte) []TXOutput {
	var UTXOs []TXOutput
	db := u.Blockchain.Db

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucket))
		c := bucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs := DeserializeOutputs(v)
			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Panicln(err)
	}
	return UTXOs
}

func (u UTXOSet) CountTransactions() int {
	db := u.Blockchain.Db
	counter := 0

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			counter++
		}
		return nil
	})
	if err != nil {
		log.Panicln(err)
	}
	return counter
}

func (u UTXOSet) Reindex() {
	db := u.Blockchain.Db
	bucketName := []byte(utxoBucket)

	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(bucketName)
		if err != nil {
			log.Panicln(err)
		}
		_, err = tx.CreateBucket(bucketName)
		if err != nil {
			log.Panicln(err)
		}
		return nil
	})
	if err != nil {
		log.Panicln(err)
	}
	UTXO := u.Blockchain.FindUTXO()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))

		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID)
			if err != nil {
				log.Panicln(err)
			}
			err = b.Put(key, outs.Serialize())
			if err != nil {
				log.Panicln(err)
			}
		}
		return nil
	})
	if err != nil {
		log.Panicln(err)
	}
}

func (u UTXOSet) Update(block *Block) {
	db := u.Blockchain.Db

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		for _, tx := range block.Transactions {
			if tx.IsCoinbase() == false {
				for _, vin := range tx.Vin {
					updateOuts := TXOutputs{}
					outsBytes := b.Get(vin.Txid)
					outs := DeserializeOutputs(outsBytes)
					for outIdx, out := range outs.Outputs {
						if outIdx != vin.Vout {
							updateOuts.Outputs = append(updateOuts.Outputs, out)
						}
					}

					if len(updateOuts.Outputs) == 0 {
						err := b.Delete(vin.Txid)
						if err != nil {
							log.Panicln(err)
						}
					} else {
						err := b.Put(vin.Txid, updateOuts.Serialize())
						if err != nil {
							log.Panicln(err)
						}
					}
				}
			}

			newOutputs := TXOutputs{}
			for _, out := range tx.Vout {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}
			err := b.Put(tx.ID, newOutputs.Serialize())
			if err != nil {
				log.Panicln(err)
			}
		}
		return nil
	})
	if err != nil {
		log.Panicln(err)
	}
}


































