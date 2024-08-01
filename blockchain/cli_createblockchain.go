package blockchain

import "log"

func (cli *CLI) createBlockchain(address string) {
	if !ValidateAddress(address) {
		log.Panicln("ERROR: address is not valid")
	}
	blockchain := CreateBlockchain(address)
	set := UTXOSet{blockchain}
	set.Reindex()
	blockchain.Db.Close()
	log.Println("Create BlockChain Done!")
}

