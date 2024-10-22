package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
)

type TXOutput struct {
	Value      int
	PubKeyHash []byte
}

// Lock sigins the output with the pubkey in address
func (out *TXOutput) Lock(address []byte) {
	pubKeyHash := Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(pubKeyHash, out.PubKeyHash) == 0
}

func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	txo.Lock([]byte(address))
	return txo
}

type TXOutputs struct {
	Outputs []TXOutput
}

func (outs TXOutputs) Serialize() []byte {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(outs)
	if err != nil {
		log.Panicln(err)
	}
	return buf.Bytes()
}

func DeserializeOutputs(data []byte) TXOutputs {
	var outputs TXOutputs
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&outputs)
	if err != nil {
		log.Panicln(err)
	}
	return outputs
}

















