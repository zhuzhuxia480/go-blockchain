package blockchain

import (
	"log"
)

func (cli *CLI) reindexUTXO() {
	bc := NewBlockChain()
	set := UTXOSet{bc}
	set.Reindex()
	count := set.CountTransactions()
	log.Printf("Done! There are %d transactions in the UTXO set.\n", count)

}