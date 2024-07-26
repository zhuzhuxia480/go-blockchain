package blockchain

import (
	"encoding/hex"
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
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

// MineBLock mines a block with the provided transactions
func (bc *BlockChain) MineBLock(transactions []*Transaction) {
	var lastHash []byte
	err := bc.Db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		lastHash = bucket.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panicln(err)
	}
	newBlock := NewBlock(transactions, lastHash)
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
}

// FindUnspentTransactions returns a list of transactions containing unspent outputs
func (bc *BlockChain) FindUnspentTransactions(address string) []Transaction {
	var unSpentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()
	for {
		block := bci.Next()
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				if out.CanBeUnlockedWith(address) {
					unSpentTXs = append(unSpentTXs, *tx)
				}
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}
		if len(block.PreBlockHash) == 0 {
			break
		}
	}
	return unSpentTXs
}

// FindUTXO finds and return unspent transactions outputs
func (bc *BlockChain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := bc.FindUnspentTransactions(address)
	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

// FindSpendableOutputs finds and returns unspent output to reference in inputs
func (bc *BlockChain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTxs := bc.FindUnspentTransactions(address)
	accumulated := 0
Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)
		for outIdx, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				if accumulated >= amount {
					break Work
				}
			}
		}
	}
	return accumulated, unspentOutputs
}

func (bc *BlockChain) Iterator() *BlockchainIterator {
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

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

func NewBlockChain(address string) *BlockChain {
	if dbExists() == false {
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

func CreateBlockchain(address string) *BlockChain {
	if dbExists() {
		log.Println("Blockchain already exists.")
		os.Exit(1)
	}
	var tip []byte
	db, err := bolt.Open(dbFile, os.ModePerm, nil)
	if err != nil {
		log.Panicln(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
		block := NewGenesisBlock(cbtx)
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
