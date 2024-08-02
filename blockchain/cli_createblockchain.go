package blockchain

import "log"

func (cli *CLI) createBlockchain(address, nodeID string) {
	if !ValidateAddress(address) {
		log.Panicln("ERROR: address is not valid")
	}
	blockchain := CreateBlockchain(address, nodeID)
	set := UTXOSet{blockchain}
	set.Reindex()
	blockchain.Db.Close()
	log.Println("Create BlockChain Done!")
}

