package blockchain

import (
	"github.com/boltdb/bolt"
	"log"
)

type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
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
