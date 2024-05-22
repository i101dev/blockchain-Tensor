package cli

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/i101dev/blockchain-Tensor/api"
)

type CommandLine struct {
}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("-  createblockchain -address ADDRESS creates a blockchain and sends genesis reward to address")
	fmt.Println("-  getbalance -address ADDRESS - get the balance for an address")
	fmt.Println("-  send -from FROM -to TO -amount AMOUNT -mine - Send amount of coins")

	fmt.Println("-  listAddresses - Lists the addresses in out wallet file")
	fmt.Println("-  createWallet - Creates a new Wallet")
	fmt.Println("-  reindexutxo - Rebuilds the UTXO set")
	fmt.Println("-  printchain - Prints the blocks in the chain")

	fmt.Println("-  startnode -miner ADDRESS - Start a node with ID specified in NODE_ID env.")
}

func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

func (cli *CommandLine) Run() {

	cli.validateArgs()

	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		log.Fatal("*** >>> NODE_ID env is not set <<< ***")
	}

	startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)

	createWalletCmd := flag.NewFlagSet("createWallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listAddresses", flag.ExitOnError)
	reindexutxoCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")
	sendMine := sendCmd.Bool("mine", false, "Mine immediately on the same node")
	startNodeMiner := startNodeCmd.String("miner", "", "Enable mining mode and send reward")

	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}

	switch os.Args[1] {

	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		Handle(err)

	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		Handle(err)

	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		Handle(err)

	case "send":
		err := sendCmd.Parse(os.Args[2:])
		Handle(err)

	case "createWallet":
		err := createWalletCmd.Parse(os.Args[2:])
		Handle(err)

	case "listAddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		Handle(err)

	case "reindexutxo":
		err := reindexutxoCmd.Parse(os.Args[2:])
		Handle(err)

	case "startnode":
		err := startNodeCmd.Parse(os.Args[2:])
		Handle(err)

	default:
		cli.printUsage()
		runtime.Goexit()
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			runtime.Goexit()
		}
		api.GetBalance(*getBalanceAddress, nodeID)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			runtime.Goexit()
		}
		api.CreateBlockchain(*createBlockchainAddress, nodeID)
	}

	if printChainCmd.Parsed() {
		api.PrintChain(nodeID)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}

		api.Send(*sendFrom, *sendTo, *sendAmount, nodeID, *sendMine)
	}

	if createWalletCmd.Parsed() {
		api.CreateWallet(nodeID)
	}

	if listAddressesCmd.Parsed() {
		api.ListAddresses(nodeID)
	}

	if reindexutxoCmd.Parsed() {
		api.ReindexUTXO(nodeID)
	}

	if startNodeCmd.Parsed() {

		if nodeID == "" {
			startNodeCmd.Usage()
			runtime.Goexit()
		}

		api.StartNode(nodeID, *startNodeMiner)
	}
}

func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}
