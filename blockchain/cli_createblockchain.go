package blockchain

import "log"

func (cli *CLI) createBlockchain(address string) {
	if !ValidateAddress(address) {
		log.Panicln("ERROR: address is not valid")
	}
	blockchain := CreateBlockchain(address)
	blockchain.Db.Close()
	set := UTXOSet{blockchain}
	set.Reindex()
	log.Println("Create BlockChain Done!")
}

