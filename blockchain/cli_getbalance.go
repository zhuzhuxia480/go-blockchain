package blockchain

import "log"

func (cli *CLI) getBalance(address string) {
	if !ValidateAddress(address) {
		log.Panicln("ERROR: address is not valid")
	}

	bc := NewBlockChain(address)
	defer bc.Db.Close()

	balance := 0
	pubKeyHash := Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressChecksumLen]
	UTXOs := bc.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}
	log.Printf("Balance of '%s' is : '%d'", address, balance)
}