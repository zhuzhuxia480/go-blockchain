package blockchain

import "log"

func (cli *CLI) send(from, to string, amount int) {
	if !ValidateAddress(from) {
		log.Panicln("ERROR: sender address is not valid")
	}
	if !ValidateAddress(to) {
		log.Panicln("ERROR: recipient address is valid")
	}

	bc := NewBlockChain(from)
	defer bc.Db.Close()

	tx := NewUTXOTransaction(from, to, amount, bc)
	bc.MineBLock([]*Transaction{tx})
	log.Println("Success!")
}
