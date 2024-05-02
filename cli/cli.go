package cli

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"

	"github.com/i101dev/golang-blockchain/blockchain"
	"github.com/i101dev/golang-blockchain/wallet"
)

type CommandLine struct {
}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("-  createblockchain -address ADDRESS creates a blockchain and sends genesis reward to address")
	fmt.Println("-  getbalance -address ADDRESS - get the balance for an address")
	fmt.Println("-  send -from FROM -to TO -amount AMOUNT - Send amount of coins")

	fmt.Println("-  listAddresses - Lists the addresses in out wallet file")
	fmt.Println("-  createWallet - Creates a new Wallet")
	fmt.Println("-  reindexutxo - Rebuilds the UTXO set")
	fmt.Println("-  printchain - Prints the blocks in the chain")
}

func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

func (cli *CommandLine) printChain() {

	chain := blockchain.ContinueBlockChain("")
	defer chain.Database.Close()
	iter := chain.Iterator()

	for {

		block := iter.NextBlock()

		pow := blockchain.NewProof(block)
		powIsValid := strconv.FormatBool(pow.Validate())

		// fmt.Printf("\n -------- \n *** >>> Block Transactions: %+v\n -------- \n", block.Transactions)
		fmt.Printf("\nPrevious hash: %x", block.PrevHash)
		fmt.Printf("\nBlock hash: %x\n", block.Hash)
		fmt.Printf("PoW: %s\n", powIsValid)

		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}

		fmt.Println()

		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *CommandLine) createBlockchain(address string) {

	if !wallet.ValidateAddress(address) {
		log.Panic("Bogus address!")
	}

	chain := blockchain.InitBlockChain(address)
	chain.Database.Close()

	UTXOset := blockchain.UTXOSet{Blockchain: chain}
	UTXOset.Reindex()

	fmt.Println("Finished creating blockchain")
}

func (cli *CommandLine) getBalance(address string) {

	if !wallet.ValidateAddress(address) {
		log.Panic("Bogus address!")
	}

	chain := blockchain.ContinueBlockChain(address)
	UTXOset := blockchain.UTXOSet{Blockchain: chain}
	defer chain.Database.Close()

	balance := 0

	pubKeyHash := wallet.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]

	UTXOs := UTXOset.FindUnspentTransactions(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of %s: %d\n", address, balance)
}

func (cli *CommandLine) send(from string, to string, amount int) {

	if !wallet.ValidateAddress(from) {
		log.Panic("Bogus - from - address!")
	}

	if !wallet.ValidateAddress(to) {
		log.Panic("Bogus - to - address!")
	}

	chain := blockchain.ContinueBlockChain(from)
	UTXOset := blockchain.UTXOSet{Blockchain: chain}
	defer chain.Database.Close()

	tx := blockchain.NewTransaction(from, to, amount, &UTXOset)
	cbtx := blockchain.CoinbaseTx(from, "")

	block := chain.AddBlock([]*blockchain.Transaction{cbtx, tx})

	UTXOset.Update(block)
	chain.AddBlock([]*blockchain.Transaction{tx})

	fmt.Println("Send transaction was a success")
}

func (cli *CommandLine) createWallet() {
	wallets, _ := wallet.CreateWallets()
	address := wallets.AddWallet()
	wallets.SaveFile()

	fmt.Printf("New address is: %s\n", address)
}

func (cli *CommandLine) listAddresses() {
	wallets, _ := wallet.CreateWallets()

	addresses := wallets.GetAllAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *CommandLine) reindexutxo() {
	chain := blockchain.ContinueBlockChain("")
	defer chain.Database.Close()
	UTXOset := blockchain.UTXOSet{Blockchain: chain}
	UTXOset.Reindex()

	count := UTXOset.CountTransactions()

	fmt.Printf("\n*** >>> Reindexing complete! There are %d transactions in the UTXO set.\n", count)
	fmt.Println()
}

func (cli *CommandLine) Run() {
	cli.validateArgs()

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

	default:
		cli.printUsage()
		runtime.Goexit()
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			runtime.Goexit()
		}
		cli.getBalance(*getBalanceAddress)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			runtime.Goexit()
		}
		cli.createBlockchain(*createBlockchainAddress)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}

		cli.send(*sendFrom, *sendTo, *sendAmount)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet()
	}
	if listAddressesCmd.Parsed() {
		cli.listAddresses()
	}
	if reindexutxoCmd.Parsed() {
		cli.reindexutxo()
	}
}

func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}
