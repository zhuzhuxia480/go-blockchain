package blockchain

import "log"

func (cli *CLI) send(from, to string, amount int, nodeID string, mineNow bool) {
	if !ValidateAddress(from) {
		log.Panicln("ERROR: sender address is not valid")
	}
	if !ValidateAddress(to) {
		log.Panicln("ERROR: recipient address is valid")
	}

	bc := NewBlockChain(nodeID)
	UTXOSet := UTXOSet{bc}
	defer bc.Db.Close()

	wallets, err := NewWallets(nodeID)
	if err != nil {
		log.Panicln(err)
	}
	wallet := wallets.GetWallet(from)
	tx := NewUTXOTransaction(&wallet, to, amount, &UTXOSet)

	if mineNow {
		coinbaseTX := NewCoinbaseTX(from, "")
		newBlock := bc.MineBLock([]*Transaction{tx, coinbaseTX})
		UTXOSet.Update(newBlock)
	} else {
		sendTx(knownNodes[0], tx)
	}
	log.Println("Success!")
}
