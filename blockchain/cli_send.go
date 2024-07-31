package blockchain

import "log"

func (cli *CLI) send(from, to string, amount int) {
	if !ValidateAddress(from) {
		log.Panicln("ERROR: sender address is not valid")
	}
	if !ValidateAddress(to) {
		log.Panicln("ERROR: recipient address is valid")
	}

	bc := NewBlockChain()
	UTXOSet := UTXOSet{bc}
	defer bc.Db.Close()

	tx := NewUTXOTransaction(from, to, amount, &UTXOSet)
	coinbaseTX := NewCoinbaseTX(from, "")
	newBlock := bc.MineBLock([]*Transaction{tx, coinbaseTX})
	UTXOSet.Update(newBlock)
	log.Println("Success!")
}
