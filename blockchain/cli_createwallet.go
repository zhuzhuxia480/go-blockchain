package blockchain

import "log"

func (cli *CLI) createWallet() {
	wallets, _ := NewWallets()
	address := wallets.CreateWallet()
	wallets.SaveToFile()
	log.Printf("You new address: %s\n", address)
}
