package blockchain

import "log"

func (cli *CLI) listAddress(nodeID string) {
	wallets, err := NewWallets(nodeID)
	if err != nil {
		log.Panicln(err)
	}
	address := wallets.GetAddresses()
	log.Println("get addresses blow:")
	for _, addr := range address {
		log.Println(addr)
	}
}
