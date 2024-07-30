package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"github.com/boltdb/bolt"
	"log"
	"os"
)

type BlockChain struct {
	tip []byte
	Db  *bolt.DB
}

const dbFile = "blockchain.Db"
const blockBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

// MineBLock mines a block with the provided transactions
func (bc *BlockChain) MineBLock(transactions []*Transaction) {
	var lastHash []byte

	for _, tx := range transactions {
		if !bc.VerifyTransaction(tx) {
			log.Panicln("ERROR: invalid transaction")
		}
	}

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

// FindUnspentTransactions returns a list of transactions containing unspent outputs
func (bc *BlockChain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
	var unSpentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()
	for {
		block := bci.Next()
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)
			log.Printf("FindUnspentTransactions tx hash: %x\n", tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							log.Printf("txID:%x used\n", tx.ID)
							continue Outputs
						}
					}
				}
				if out.IsLockedWithKey(pubKeyHash) {
					log.Printf("txID:%x not used\n", tx.ID)
					unSpentTXs = append(unSpentTXs, *tx)
				}
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.UsesKey(pubKeyHash) {
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
func (bc *BlockChain) FindUTXO(FindUTXO []byte) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := bc.FindUnspentTransactions(FindUTXO)
	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.IsLockedWithKey(FindUTXO) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

// FindSpendableOutputs finds and returns unspent output to reference in inputs
func (bc *BlockChain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTxs := bc.FindUnspentTransactions(pubKeyHash)
	accumulated := 0
Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)
		for outIdx, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
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
