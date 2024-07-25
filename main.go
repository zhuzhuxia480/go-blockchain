package main

import "go-blockchain/blockchain"

func main() {
	bc := blockchain.NewBlockChain()
	defer bc.Db.Close()
	cli := blockchain.CLI{Bc: bc}
	cli.Run()
}
