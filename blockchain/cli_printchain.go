package blockchain

import (
	"log"
	"strconv"
)

func (cli *CLI) printChain() {
	bc := NewBlockChain()
	defer bc.Db.Close()
	iterator := bc.Iterator()
	for {
		block := iterator.Next()
		log.Printf("============= Block %x =============\n", block.Hash)
		log.Printf("Prev. block.hash: %x\n", block.PreBlockHash)
		pow := NewProofOfWork(block)
		log.Println("Pow: ", strconv.FormatBool(pow.Validate()))
		for _, tx := range block.Transactions {
			log.Println(tx)
		}
		log.Println()
		if len(block.PreBlockHash) == 0 {
			log.Println("iterator chain completed")
			break
		}
	}
}
