package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/i101dev/blockchain-Tensor/blockchain"
	"github.com/i101dev/blockchain-Tensor/network"
	"github.com/i101dev/blockchain-Tensor/wallet"
)

func HandleCreateBlockchain(w http.ResponseWriter, r *http.Request) {
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

func HandleGetBalance(w http.ResponseWriter, r *http.Request) {
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

func HandleSend(w http.ResponseWriter, r *http.Request) {
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
	success, err := Send(SendRequest.From, SendRequest.To, SendRequest.Amount, nodeID, SendRequest.Mine)

	w.WriteHeader(http.StatusOK)

	if success {
		fmt.Fprintln(w, "Send transaction was a success")
	} else {
		fmt.Fprintln(w, "Send transaction was a FAILURE -", err)
	}
}

func HandleStartNode(w http.ResponseWriter, r *http.Request) {
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

func HandleListAddresses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	nodeID := os.Getenv("NODE_ID")
	addresses := ListAddresses(nodeID)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(addresses)
}

func HandleCreateWallet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	nodeID := os.Getenv("NODE_ID")
	address := CreateWallet(nodeID)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "New address is: %s\n", address)
}

func HandleReindexUTXO(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	nodeID := os.Getenv("NODE_ID")
	ReindexUTXO(nodeID)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Reindexing complete")
}

func HandlePrintChain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	nodeID := os.Getenv("NODE_ID")
	chainData := PrintChain(nodeID)

	w.WriteHeader(http.StatusOK)
	w.Write(chainData)
}

func HandleListNodes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	chainData := network.KnownNodes
	log.Printf("Number of known nodes: %d", len(chainData))

	jsonData, err := json.Marshal(chainData)
	if err != nil {
		http.Error(w, "Failed to encode data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, writeErr := w.Write(jsonData)
	if writeErr != nil {
		log.Printf("Failed to write response: %v", writeErr)
	}
}

// -------------------------------------------------------------------------------
func CreateBlockchain(address string, nodeID string) {
	if !wallet.ValidAddress(address) {
		log.Panic("Bogus address!")
	}

	chain := blockchain.InitBlockChain(address, nodeID)
	defer chain.Database.Close()

	UTXOset := blockchain.UTXOSet{Blockchain: chain}
	UTXOset.Reindex()

	fmt.Println("Finished creating blockchain")
}

func GetBalance(address string, nodeID string) int {
	if !wallet.ValidAddress(address) {
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

	fmt.Println("Balance: ", balance)

	return balance
}

func Send(from string, to string, amount int, nodeID string, mineNow bool) (bool, error) {

	if len(network.KnownNodes) < 1 {
		return false, fmt.Errorf("zero nodes online")
	}

	if !wallet.ValidAddress(from) {
		log.Panic("\n*** >>> Bogus - from - address!")
	}

	if !wallet.ValidAddress(to) {
		log.Panic("\n*** >>> Bogus - to - address!")
	}

	chain := blockchain.ContinueBlockChain(nodeID)
	UTXOset := blockchain.UTXOSet{Blockchain: chain}
	defer chain.Database.Close()

	wallets, err := wallet.LoadWallet(nodeID)
	if err != nil {
		log.Panicln("\n*** >>> error creating wallets", err)
	}

	wallet := wallets.GetAccount(from)

	tx, success := blockchain.NewTransaction(&wallet, from, to, amount, &UTXOset)

	if success {

		if mineNow {
			cbtx := blockchain.CoinbaseTx(from, "")
			txs := []*blockchain.Transaction{cbtx, tx}
			block := chain.MineBlock(txs)
			UTXOset.Update(block)
		} else {
			err = network.SendTx(network.KnownNodes[0], tx)
			return err == nil, err
			// if err != nil {
			// 	return false, err
			// } else {
			// 	return true, nil
			// }
		}
	}

	return success, nil
}

func CreateWallet(nodeID string) string {
	wallets, _ := wallet.LoadWallet(nodeID)
	address := wallets.AddAccount()
	wallets.SaveFile(nodeID)

	return address
}

func ListAddresses(nodeID string) []string {
	wallets, _ := wallet.LoadWallet(nodeID)
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

		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Prev. hash: %x\n", block.PrevHash)
		pow := blockchain.NewProof(block)
		powIsValid := strconv.FormatBool(pow.Validate())

		blockInfo := fmt.Sprintf("\nPrevious hash: %x\nBlock hash: %x\nPoW: %s\n", block.PrevHash, block.Hash, powIsValid)
		result = append(result, []byte(blockInfo)...)

		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range block.Transactions {
			txInfo := fmt.Sprintf("%+v\n", tx)
			result = append(result, []byte(txInfo)...)
			fmt.Println(tx)
		}
		fmt.Println()

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return result
}

func StartNode(nodeID string, minerAddress string) {
	fmt.Printf("Starting Node %s\n", nodeID)

	if len(minerAddress) > 0 {
		if wallet.ValidAddress(minerAddress) {
			fmt.Println("\n*** >>> Mining is on. Address to receive rewards: ", minerAddress)
		} else {
			log.Panic("\n*** >>> Wrong miner address")
		}
	}

	network.StartServer(nodeID, minerAddress)
}
