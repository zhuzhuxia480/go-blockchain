package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"os"
)

type BlockChain struct {
	tip []byte
	Db  *bolt.DB
}

const dbFile = "blockchain_%s.db"
const blockBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

// MineBLock mines a block with the provided transactions
func (bc *BlockChain) MineBLock(transactions []*Transaction) *Block {
	var lastHash []byte
	var lastHeight int

	for _, tx := range transactions {
		if !bc.VerifyTransaction(tx) {
			log.Panicln("ERROR: invalid transaction")
		}
	}

	err := bc.Db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		lastHash = bucket.Get([]byte("l"))
		lastHeight = DeSerializeBlock(bucket.Get(lastHash)).Height
		return nil
	})
	if err != nil {
		log.Panicln(err)
	}
	newBlock := NewBlock(transactions, lastHash, lastHeight+1)
	err = bc.Db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		err := bucket.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panicln(err)
		}
		err = bucket.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panicln(err)
		}
		bc.tip = newBlock.Hash
		return nil
	})
	if err != nil {
		log.Panicln(err)
	}
	return newBlock
}

func (bc *BlockChain) AddBlock(block *Block) {
	err := bc.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockBucket))
		blockInDb := b.Get(block.Hash)
		if blockInDb != nil {
			return nil
		}

		blockData := block.Serialize()
		err := b.Put(block.Hash, blockData)
		if err != nil {
			log.Panicln(err)
		}

		lastHash := b.Get([]byte("l"))
		lastBlockData := b.Get(lastHash)
		lastBlock := DeSerializeBlock(lastBlockData)
		if block.Height > lastBlock.Height {
			err := b.Put([]byte("l"), block.Hash)
			if err != nil {
				log.Panicln(err)
			}
			bc.tip = block.Hash
		}
		return nil
	})
	if err != nil {
		log.Panicln(err)
	}
}

func (bc *BlockChain) GetBestHeight() int {
	var lastBlock Block
	err := bc.Db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		lastHash := bucket.Get([]byte("l"))
		lastBlock = *DeSerializeBlock(bucket.Get(lastHash))
		return nil
	})
	if err != nil {
		log.Panicln(err)
	}
	return lastBlock.Height
}

func (bc *BlockChain) GetBlock(blockHash []byte) (Block, error) {
	var block Block

	err := bc.Db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		blockData := bucket.Get(blockHash)
		if blockData == nil {
			return errors.New("block is not found")
		}
		block = *DeSerializeBlock(blockData)
		return nil
	})
	if err != nil {
		return block, err
	}
	return block, nil
}

func (bc *BlockChain) GetBlockHashes() [][]byte {
	var blocks [][]byte
	bci := bc.Iterator()

	for {
		block := bci.Next()
		blocks = append(blocks, block.Hash)
		if len(block.PreBlockHash) == 0 {
			break
		}
	}
	return blocks
}

func (bc *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	preTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		preTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panicln(err)
		}
		preTXs[hex.EncodeToString(preTX.ID)] = preTX
	}
	tx.Sign(privKey, preTXs)
}

// FindUTXO finds and return unspent transactions outputs
func (bc *BlockChain) FindUTXO() map[string]TXOutputs {
	UTXO := make(map[string]TXOutputs)
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}
				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}
		}

		if len(block.PreBlockHash) == 0 {
			break
		}
	}
	return UTXO
}

func (bc *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	iterator := bc.Iterator()
	for {
		block := iterator.Next()
		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}
		if len(block.PreBlockHash) == 0 {
			break
		}
	}
	return Transaction{}, errors.New("transaction is not found")
}

func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}
	prevTXs := make(map[string]Transaction)
	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panicln(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	return tx.Verify(prevTXs)
}

func (bc *BlockChain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{bc.tip, bc.Db}
}

func dbExists(dbFile string) bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

func NewBlockChain(nodeID string) *BlockChain {
	dbFile := fmt.Sprintf(dbFile, nodeID)
	if dbExists(dbFile) == false {
		log.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}
	var tip []byte
	db, err := bolt.Open(dbFile, os.ModePerm, nil)
	if err != nil {
		log.Panicln("err:", err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockBucket))
		tip = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panicln("Db update err:", err)
	}
	return &BlockChain{tip, db}
}

func CreateBlockchain(address string, nodeID string) *BlockChain {
	dbFile := fmt.Sprintf(dbFile, nodeID)
	if dbExists(dbFile) {
		log.Println("Blockchain already exists.")
		os.Exit(1)
	}
	var tip []byte
	cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
	block := NewGenesisBlock(cbtx)

	db, err := bolt.Open(dbFile, os.ModePerm, nil)
	if err != nil {
		log.Panicln(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucket([]byte(blockBucket))
		if err != nil {
			log.Panicln(err)
		}
		err = bucket.Put(block.Hash, block.Serialize())
		if err != nil {
			log.Panicln(err)
		}
		err = bucket.Put([]byte("l"), block.Hash)
		if err != nil {
			log.Panicln(err)
		}
		tip = block.Hash
		return nil
	})
	if err != nil {
		log.Panicln(err)
	}
	return &BlockChain{
		tip: tip,
		Db:  db,
	}
}
