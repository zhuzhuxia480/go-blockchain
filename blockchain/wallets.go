package blockchain

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"os"
)

const walletFile = "wallet_%s.dat"

type Wallets struct {
	Wallets map[string]*Wallet
}

func NewWallets(nodeID string) (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)
	err := wallets.LoadFromFile(nodeID)
	return &wallets, err
}

func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet()
	address := wallet.GetAddress()
	log.Printf("get new address:%s", address)
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

func (ws *Wallets) LoadFromFile(nodeID string) error {
	walletFile := fmt.Sprintf(walletFile, nodeID)
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	fileContent, err := os.ReadFile(walletFile)
	if err != nil {
		log.Panicln(err)
	}

	var wallets Wallets
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil {
		log.Panicln(err)
	}
	if err != nil {
		log.Panicln(err)
	}
	ws.Wallets = wallets.Wallets
	return nil
}

func (ws Wallets) SaveToFile(nodeID string) {
	var content bytes.Buffer
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(&ws)
	if err != nil {
		log.Panic(err)
	}
	walletFile := fmt.Sprintf(walletFile, nodeID)
	err = os.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}









