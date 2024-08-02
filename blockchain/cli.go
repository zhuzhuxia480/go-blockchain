package blockchain

import (
	"flag"
	"log"
	"os"
)

type CLI struct {
}

func (cli *CLI) validArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) printUsage() {
	log.Println("Usage:")
	log.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	log.Println("  createwallet - Generates a new key-pair and saves it into the wallet file")
	log.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	log.Println("  listaddresses - Lists all addresses from the wallet file")
	log.Println("  printchain - Print all the blocks of the blockchain")
	log.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
	log.Println("  send -from FROM -to TO -amount AMOUNT -mine - Send AMOUNT of coins from FROM address to TO")
	log.Println("  startnode -miner ADDRESS - Start a node with ID specified in NODE_ID env. -miner enables mining")
}

func (cli *CLI) Run() {
	cli.validArgs()

	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		log.Panicln("NODE_ID env is not set")
	}

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "the address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "the address to send genesis reward to")
	sendTo := sendCmd.String("to", "", "destination wallet address")
	sendFrom := sendCmd.String("from", "", "source wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")
	sendMine := startNodeCmd.Bool("mine", false, "Mine immediately on the same node")
	startNodeMiner := startNodeCmd.String("miner", "", "")

	switch os.Args[1] {
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panicln(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panicln(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panicln(err)
		}
	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panicln(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panicln(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panicln(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet()
	}

	if listAddressesCmd.Parsed() {
		cli.listAddress()
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if sendCmd.Parsed() {
		if *sendTo == "" || *sendFrom == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}
		cli.send(*sendFrom, *sendTo, *sendAmount)
	}
}
