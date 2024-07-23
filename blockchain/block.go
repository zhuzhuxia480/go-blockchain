package blockchain

import (
	"bytes"
	"crypto/sha256"
	"strconv"
	"time"
)

type Block struct {
	Timestamp    int64
	Data         []byte
	PreBlockHash []byte
	Hash         []byte
	Nonce        int
}

func (b *Block) SetHash() {
	timeStamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	headers := bytes.Join([][]byte{b.PreBlockHash, b.Data, timeStamp}, []byte{})
	hash := sha256.Sum256(headers)
	b.Hash = hash[:]
}

func NewBlock(data string, preBlockHash []byte) *Block {
	block := &Block{
		time.Now().Unix(),
		[]byte(data),
		preBlockHash,
		[]byte{},
		0,
	}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()
	block.Hash = hash
	block.Nonce = nonce
	return block
}

func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}
