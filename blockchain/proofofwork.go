package proofofwork

import (
	"bytes"
	"crypto/sha256"
	"go-blockchain/block"
	"go-blockchain/util"
	"log"
	"math"
	"math/big"
)

const targetBit = 24
const maxNonce = math.MaxInt64

type ProofOfWork struct {
	block  *block.Block
	target *big.Int
}

func NewProofOfWork(b *block.Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBit))
	return &ProofOfWork{
		block:  b,
		target: target,
	}
}

func (pow ProofOfWork) PrepareData(nonce int) []byte {
	bytes := bytes.Join([][]byte{
		pow.block.PreBlockHash,
		pow.block.Data,
		util.IntToHex(pow.block.Timestamp),
		util.IntToHex(int64(targetBit)),
		util.IntToHex(int64(nonce)),
	}, []byte{})
	return bytes
}



func (pow ProofOfWork) Run() (int, []byte) {
	nonce := 0
	var hash [32]byte
	var hasInt big.Int
	for nonce = 0; nonce < maxNonce; nonce++ {
		data := pow.PrepareData(nonce)
		hash = sha256.Sum256(data)
		log.Printf("\r%x", hash)
		hasInt.SetBytes(hash[:])
		if hasInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}
	return nonce, hash[:]
}

func (pow ProofOfWork) Validate() bool {
	var hashInt big.Int
	data := pow.PrepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])
	if hashInt.Cmp(pow.target) == -1 {
		return true
	}
	return false
}