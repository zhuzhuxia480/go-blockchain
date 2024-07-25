package blockchain

import (
	"flag"
	"log"
	"os"
	"strconv"
)

type CLI struct {
	Bc *BlockChain
}

func (cli *CLI) printUsage() {
	log.Println("Usage:")
	log.Println("addblock --data BLOCK_DATA - add a block to the blockchain")
	log.Println("printchain - print all the blocks  of the blockchain")
}

func (cli *CLI) validArgs() {
	if len(os.Args) < 2{
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) addBlock(data string) {
	cli.Bc.AddBlock(data)
}

func (cli *CLI) printChain() {
	iterator := cli.Bc.Iterator()
	for  {
		block := iterator.Next()
		log.Printf("Pre hash: %x\n", block.PreBlockHash)
		log.Println("Data: ", string(block.Data))
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









































