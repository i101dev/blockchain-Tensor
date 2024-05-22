package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"

	"github.com/i101dev/blockchain-Tensor/blockchain"
	"github.com/i101dev/blockchain-Tensor/network"
	"github.com/i101dev/blockchain-Tensor/wallet"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file")
	}

}

func main() {
	defer os.Exit(0)
	// c := cli.CommandLine{}
	// c.Run()

	http.HandleFunc("/createblockchain", handleCreateBlockchain)
	http.HandleFunc("/getbalance", handleGetBalance)
	http.HandleFunc("/send", handleSend)
	http.HandleFunc("/startnode", handleStartNode)
	http.HandleFunc("/listaddresses", handleListAddresses)
	http.HandleFunc("/createwallet", handleCreateWallet)
	http.HandleFunc("/reindexutxo", handleReindexUTXO)
	http.HandleFunc("/printchain", handlePrintChain)

	nodeID := os.Getenv("NODE_ID")
	port := os.Getenv("PORT")

	if nodeID == "" {
		log.Fatal("*** >>> NODE_ID env is not set <<< ***")
		runtime.Goexit()
	}

	log.SetFlags(0)
	log.Printf("Server started on Port: %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleCreateBlockchain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var CreateBlockchainRequest struct {
		Address string `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&CreateBlockchainRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	nodeID := os.Getenv("NODE_ID")
	CreateBlockchain(CreateBlockchainRequest.Address, nodeID)

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Blockchain created and reward sent to address %s\n", CreateBlockchainRequest.Address)
}

func handleGetBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var GetBalanceRequest struct {
		Address string `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&GetBalanceRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	nodeID := os.Getenv("NODE_ID")
	balance := GetBalance(GetBalanceRequest.Address, nodeID)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Balance of %s: %d\n", GetBalanceRequest.Address, balance)
}

func handleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var SendRequest struct {
		From   string `json:"from"`
		To     string `json:"to"`
		Amount int    `json:"amount"`
		Mine   bool   `json:"mine"`
	}
	if err := json.NewDecoder(r.Body).Decode(&SendRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	nodeID := os.Getenv("NODE_ID")
	Send(SendRequest.From, SendRequest.To, SendRequest.Amount, nodeID, SendRequest.Mine)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Send transaction was a success")
}

func handleStartNode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var startNodeRequest struct {
		MinerAddress string `json:"miner_addr"`
	}
	if err := json.NewDecoder(r.Body).Decode(&startNodeRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	nodeID := os.Getenv("NODE_ID")
	StartNode(nodeID, startNodeRequest.MinerAddress)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Node %s started\n", nodeID)
}

func handleListAddresses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	nodeID := os.Getenv("NODE_ID")
	addresses := ListAddresses(nodeID)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(addresses)
}

func handleCreateWallet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	nodeID := os.Getenv("NODE_ID")
	address := CreateWallet(nodeID)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "New address is: %s\n", address)
}

func handleReindexUTXO(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	nodeID := os.Getenv("NODE_ID")
	ReindexUTXO(nodeID)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Reindexing complete")
}

func handlePrintChain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	nodeID := os.Getenv("NODE_ID")
	chainData := PrintChain(nodeID)

	w.WriteHeader(http.StatusOK)
	w.Write(chainData)
}

// -------------------------------------------------------------------------------
func CreateBlockchain(address string, nodeID string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("Bogus address!")
	}

	chain := blockchain.InitBlockChain(address, nodeID)
	defer chain.Database.Close()

	UTXOset := blockchain.UTXOSet{Blockchain: chain}
	UTXOset.Reindex()

	fmt.Println("Finished creating blockchain")
}

func GetBalance(address string, nodeID string) int {
	if !wallet.ValidateAddress(address) {
		log.Panic("Bogus address!")
	}

	chain := blockchain.ContinueBlockChain(nodeID)
	UTXOset := blockchain.UTXOSet{Blockchain: chain}
	defer chain.Database.Close()

	balance := 0

	pubKeyHash := wallet.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]

	UTXOs := UTXOset.FindUnspentTransactions(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	return balance
}

func Send(from string, to string, amount int, nodeID string, mineNow bool) {
	if !wallet.ValidateAddress(from) {
		log.Panic("Bogus - from - address!")
	}

	if !wallet.ValidateAddress(to) {
		log.Panic("Bogus - to - address!")
	}

	chain := blockchain.ContinueBlockChain(nodeID)
	UTXOset := blockchain.UTXOSet{Blockchain: chain}
	defer chain.Database.Close()

	wallets, err := wallet.CreateWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}

	wallet := wallets.GetWallet(from)

	tx := blockchain.NewTransaction(&wallet, from, to, amount, &UTXOset)

	if mineNow {
		cbtx := blockchain.CoinbaseTx(from, "")
		txs := []*blockchain.Transaction{cbtx, tx}
		block := chain.MineBlock(txs)
		UTXOset.Update(block)
	} else {
		network.SendTx(network.KnownNodes[0], tx)
		fmt.Println("send tx")
	}

	fmt.Println()
	fmt.Println("*** >>> Send transaction was a success")
	fmt.Print()
}

func CreateWallet(nodeID string) string {
	wallets, _ := wallet.CreateWallets(nodeID)
	address := wallets.AddWallet()

	wallets.SaveFile(nodeID)

	return address
}

func ListAddresses(nodeID string) []string {
	wallets, _ := wallet.CreateWallets(nodeID)
	return wallets.GetAllAddresses()
}

func ReindexUTXO(nodeID string) {
	chain := blockchain.ContinueBlockChain(nodeID)
	defer chain.Database.Close()

	UTXOset := blockchain.UTXOSet{Blockchain: chain}
	UTXOset.Reindex()
}

func PrintChain(nodeID string) []byte {
	chain := blockchain.ContinueBlockChain(nodeID)
	defer chain.Database.Close()
	iter := chain.Iterator()

	var result []byte

	for {
		block := iter.NextBlock()
		pow := blockchain.NewProof(block)
		powIsValid := strconv.FormatBool(pow.Validate())

		blockInfo := fmt.Sprintf("\nPrevious hash: %x\nBlock hash: %x\nPoW: %s\n", block.PrevHash, block.Hash, powIsValid)
		result = append(result, []byte(blockInfo)...)

		for _, tx := range block.Transactions {
			txInfo := fmt.Sprintf("%+v\n", tx)
			result = append(result, []byte(txInfo)...)
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return result
}

func StartNode(nodeID string, minerAddress string) {
	fmt.Printf("Starting Node %s\n", nodeID)

	if len(minerAddress) > 0 {
		if wallet.ValidateAddress(minerAddress) {
			fmt.Println("Mining is on. Address to receive rewards: ", minerAddress)
		} else {
			log.Panic("Wrong miner address")
		}
	}

	network.StartServer(nodeID, minerAddress)
}
