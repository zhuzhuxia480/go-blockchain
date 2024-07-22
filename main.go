package main

import (
	"log"
)
import "go-blockchain/blockchain"

func main() {
	bc := blockchain.NewBlockChain()

	bc.AddBlock("Send 1 BTC to Ivan")
	bc.AddBlock("Send 2 more BTC to Ivan")

	for _, block := range bc.Blocks {
		log.Printf("Prev. hash: %x\n", block.PreBlockHash)
		log.Printf("Data: %s\n", block.Data)
		log.Printf("Hash: %x\n", block.Hash)
		log.Println()
	}
}
