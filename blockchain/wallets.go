package blockchain

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"log"
	"os"
)

type Wallets struct {
	Wallets map[string]*Wallet
}

func NewWallets() (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)
	err := wallets.LoadFromFile()
	return &wallets, err
}

func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet()
	address := wallet.GetAddress()
	//TODO judge string([]byte) whether equal to fmt.Sprintf("%s", []byte)
	ws.Wallets[string(address)] = wallet
	return string(address)
}

func (ws *Wallets) GetAddresses() []string {
	var addresses []string
	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}
	return addresses
}

func (ws Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

func (ws *Wallets) LoadFromFile() error {
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	fileContent, err := os.ReadFile(walletFile)
	if err != nil {
		log.Panicln(err)
	}

	var wallets Wallets
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil {
		log.Panicln(err)
	}
	ws.Wallets = wallets.Wallets
	return nil
}

func (ws Wallets) SaveToFile() {
	var content bytes.Buffer
	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		log.Panicln(err)
	}
	err = os.WriteFile(walletFile, content.Bytes(), os.ModePerm)
	if err != nil {
		log.Panicln(err)
	}
}










