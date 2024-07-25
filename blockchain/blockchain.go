package blockchain

import (
	"github.com/boltdb/bolt"
	"log"
	"os"
)

type BlockChain struct {
	tip []byte
	Db  *bolt.DB
}

type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

const dbFile = "blockchain.Db"
const blockBucket = "blocks"

func (bc *BlockChain)Iterator() *BlockchainIterator {
	return &BlockchainIterator{bc.tip, bc.Db}
}

func (it *BlockchainIterator) Next() *Block {
	var block *Block
	err := it.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		obj := bucket.Get(it.currentHash)
		block = DeSerialize(obj)
		return nil
	})
	if err != nil {
		log.Panicln(err)
	}
	it.currentHash = block.PreBlockHash
	return block
}

func (bc *BlockChain) AddBlock(data string) {
	var lastHash []byte
	err := bc.Db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		lastHash = bucket.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panicln("err:", err)
	}
	block := NewBlock(data, lastHash)

	err = bc.Db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		err := bucket.Put([]byte("l"), block.Hash)
		if err != nil {
			return err
		}
		err = bucket.Put(block.Hash, block.Serialize())
		if err != nil {
			return err
		}
		bc.tip = block.Hash
		return nil
	})
	if err != nil {
		log.Panicln(err)
	}

}

func NewBlockChain() *BlockChain {
	var tip []byte
	db, err := bolt.Open(dbFile, os.ModePerm, nil)
	if err != nil {
		log.Panicln("err:", err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockBucket))
		if b == nil {
			log.Println("this is no exist blockchain in Db, create a new one")
			b, err = tx.CreateBucket([]byte(blockBucket))
			if err != nil {
				return err
			}
			block := NewGenesisBlock()
			err := b.Put(block.Hash, block.Serialize())
			if err != nil {
				return err
			}
			err = b.Put([]byte("l"), block.Hash)
			if err != nil {
				return err
			}
			tip = block.Hash
		} else {
			tip = b.Get([]byte("l"))
		}
		return nil
	})
	if err != nil {
		log.Panicln("Db update err:", err)
	}
	return &BlockChain{tip, db}
}
