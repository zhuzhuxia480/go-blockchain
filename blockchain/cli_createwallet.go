package blockchain

import "log"

func (cli *CLI) createWallet() {
	wallets, err := NewWallets()
	if err != nil {
		log.Panicln(err)
	}
	address := wallets.CreateWallet()
	wallets.SaveToFile()
	log.Printf("You new address: %s\n", address)
}
