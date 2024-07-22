package blockchain

import "go-blockchain/block"

type BlockChain struct {
	blocks []*block.Block
}

func (bc BlockChain) AddBlock(data string) {
	prevBlock := bc.blocks[len(bc.blocks)-1]
	newBlock := block.NewBlock(data, prevBlock.Hash)
	bc.blocks = append(bc.blocks, newBlock)
}

func NewBlockChain() *BlockChain {
	return &BlockChain{[]*block.Block{block.NewGenesisBlock()}}
}
