package blockchain

import (
	"flag"
	"log"
	"os"
	"strconv"
)

type CLI struct {
}

func (cli *CLI) createBlockchain(address string) {
	blockchain := CreateBlockchain(address)
	blockchain.Db.Close()
	log.Println("Create BlockChain Done!")
}

func (cli *CLI) getBalance(address string) {
	bc := NewBlockChain(address)
	defer bc.Db.Close()

	balance := 0
	UTXOs := bc.FindUTXO(address)

	for _, out := range UTXOs {
		balance += out.Value
	}
	log.Printf("Balance of '%s' is : '%d'", address, balance)
}

func (cli *CLI) printUsage() {
	log.Println("Usage:")
	log.Println("  getbalance --address ADDRESS - Get balance of ADDRESS")
	log.Println("  createblockchain --address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	log.Println("  printchain - Print all the blocks  of the blockchain")
	log.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM to TO")
}

func (cli *CLI) validArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}


func (cli *CLI) printChain() {
	bc := NewBlockChain("")
	defer bc.Db.Close()
	iterator := bc.Iterator()
	for {
		block := iterator.Next()
		log.Printf("Pre hash: %x\n", block.PreBlockHash)
		log.Printf("Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		log.Println("Pow: ", strconv.FormatBool(pow.Validate()))
		log.Println()
		if len(block.PreBlockHash) == 0 {
			log.Println("iterator chain completed")
			break
		}
	}
}

func (cli *CLI) send(from, to string, amount int) {
	bc := NewBlockChain(from)
	defer bc.Db.Close()

	tx := NewUTXOTransaction(from, to, amount, bc)
	bc.MineBLock([]*Transaction{tx})
	log.Println("Success!")
}

func (cli *CLI) Run() {
	cli.validArgs()
	addblockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	addblockData := addblockCmd.String("data", "", "block data")
	switch os.Args[1] {
	case "addblock":
		err := addblockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panicln(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panicln(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}
	if addblockCmd.Parsed() {
		if *addblockData == "" {
			addblockCmd.Usage()
			os.Exit(1)
		}
		cli.addBlock(*addblockData)
	}
	if printChainCmd.Parsed() {
		cli.printChain()
	}
}
