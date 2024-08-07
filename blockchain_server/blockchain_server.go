package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/i101dev/blockchain-Tensor/blockchain"
	"github.com/i101dev/blockchain-Tensor/network"
	"github.com/i101dev/blockchain-Tensor/types"
	"github.com/i101dev/blockchain-Tensor/wallet"
)

var (
	cache map[string]*blockchain.Blockchain = make(map[string]*blockchain.Blockchain)

	CHAIN_ID       = "blockchain"
	ORIGIN_ADDRESS = "1CdnbM5PaWJRWMcMghkCoNPQaURHRsxFtj"
	MINER_ADDRESS  = "1JFtRuBGZDkr8rZ1kDrV6T5QZk3rmjS2Ed"
)

type BlockchainServer struct {
	port uint16
}

func (bcs *BlockchainServer) LoadBlockchain() error {

	bc, err := blockchain.LoadBlockchain(ORIGIN_ADDRESS, bcs.port)

	if err != nil {
		return fmt.Errorf("failed to load chain")
	}

	cache[CHAIN_ID] = bc

	return nil
}

func (bcs *BlockchainServer) GetBlockchain() (*blockchain.Blockchain, error) {

	bc, ok := cache[CHAIN_ID]

	if !ok {
		return nil, fmt.Errorf("failed to fetch chain data - initialization required")
	}

	return bc, nil
}

func (bcs *BlockchainServer) PrintChain(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:

		w.Header().Add("Content-Type", "application/json")

		// ----------------------------------------------------------
		bc, err := bcs.GetBlockchain()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		blockchain.OpenDB(bc)
		defer bc.CloseDB()

		allBlocks := bc.GetAllBlocks()

		// ----------------------------------------------------------
		allBlocksJSON, err := json.Marshal(allBlocks)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// ----------------------------------------------------------
		w.Header().Add("Content-Type", "application/json")
		w.Write(allBlocksJSON)

	default:
		http.Error(w, "ERROR: Invalid HTTP Method", http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) NewAccount(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:

		w.Header().Add("Content-Type", "application/json")

		// ----------------------------------------------------------
		w, _ := wallet.CreateWallets()

		w.AddAccount()

		w.SaveFile()

		w.Print()
		// ----------------------------------------------------------
		// w.Header().Add("Content-Type", "application/json")

	default:
		http.Error(w, "ERROR: Invalid HTTP Method", http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) LoadWallet(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:

		w.Header().Add("Content-Type", "application/json")

		// ----------------------------------------------------------
		w, _ := wallet.CreateWallets()

		w.Print()
		// ----------------------------------------------------------
		// w.Header().Add("Content-Type", "application/json")

	default:
		http.Error(w, "ERROR: Invalid HTTP Method", http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) GetBlock(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		hash := req.URL.Query().Get("hash")

		// ----------------------------------------------------------
		hashBytes, err := hex.DecodeString(hash)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// ----------------------------------------------------------
		bc, err := bcs.GetBlockchain()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		db := blockchain.OpenDB(bc)
		defer bc.CloseDB()

		// ----------------------------------------------------------
		block, err := bc.GetBlockByHash(db, hashBytes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// ----------------------------------------------------------
		m, err := block.MarshalJSON()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.Write(m)

	default:
		http.Error(w, "ERROR: Invalid HTTP Method", http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) GetTXN(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:

		w.Header().Add("Content-Type", "application/json")

		ID := req.URL.Query().Get("id")

		// -----------------------------------------------------------
		chain, err := bcs.GetBlockchain()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		blockchain.OpenDB(chain)
		defer chain.CloseDB()

		// -----------------------------------------------------------
		txnID, err := hex.DecodeString(ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		transaction, err := chain.FindTransaction(txnID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// -----------------------------------------------------------
		TXN, err := json.Marshal(transaction)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// -----------------------------------------------------------
		w.Header().Add("Content-Type", "application/json")
		w.Write(TXN)

	default:
		http.Error(w, "ERROR: Invalid HTTP Method", http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) AddTXN(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:

		var txnPayload types.NewTxnReq
		decoder := json.NewDecoder(req.Body)

		// ----------------------------------------------------------
		err := decoder.Decode(&txnPayload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// ----------------------------------------------------------
		chain, err := bcs.GetBlockchain()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		blockchain.OpenDB(chain)
		defer chain.CloseDB()

		// ----------------------------------------------------------
		UTXOset := blockchain.UTXOSet{
			Blockchain: chain,
		}

		wallet, err := wallet.CreateWallets()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// ----------------------------------------------------------
		newTxn := blockchain.NewTransaction(txnPayload.From, txnPayload.To, txnPayload.Amount, &UTXOset, wallet)

		if txnPayload.MineNow {
			cbTx := blockchain.CoinbaseTX(txnPayload.From, "")
			txs := []*blockchain.Transaction{cbTx, newTxn}
			block := chain.MineBlock(txs)
			UTXOset.Update(block)
		} else {
			network.SendTx(network.NODE_ZERO, newTxn)
			fmt.Println("\nsending txn")
		}

		// ----------------------------------------------------------
		m, err := newTxn.MarshalJSON()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.Write(m)

	default:
		http.Error(w, "ERROR: Invalid HTTP Method", http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) GetUTXOset(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:

		w.Header().Add("Content-Type", "application/json")

		address := req.URL.Query().Get("address")

		// -----------------------------------------------------------
		chain, err := bcs.GetBlockchain()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		db := blockchain.OpenDB(chain)
		defer chain.CloseDB()

		// -----------------------------------------------------------
		utxoset, err := chain.GetUnspentOutputs(db, address)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// -----------------------------------------------------------
		UTXO, err := json.Marshal(utxoset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// -----------------------------------------------------------
		w.Header().Add("Content-Type", "application/json")
		w.Write(UTXO)

	default:
		http.Error(w, "ERROR: Invalid HTTP Method", http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) GetBalance(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:

		w.Header().Add("Content-Type", "application/json")

		address := req.URL.Query().Get("address")

		// -----------------------------------------------------------
		chain, err := bcs.GetBlockchain()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		blockchain.OpenDB(chain)
		defer chain.CloseDB()

		UTXOset := blockchain.UTXOSet{
			Blockchain: chain,
		}

		// -----------------------------------------------------------
		walletDat, _ := wallet.CreateWallets()
		account := walletDat.GetAccount(address)
		pubKeyHash := wallet.PublicKeyHash(account.PublicKey)
		UTXOs := UTXOset.FindUnspentTransactions(pubKeyHash)

		balance := 0
		for _, out := range UTXOs {
			balance += out.Value
		}

		// -----------------------------------------------------------
		response := map[string]int{"balance": balance}
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.Write(jsonResponse)

	default:
		http.Error(w, "ERROR: Invalid HTTP Method", http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) Reindex(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:

		w.Header().Add("Content-Type", "application/json")

		// -----------------------------------------------------------
		chain, err := bcs.GetBlockchain()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		blockchain.OpenDB(chain)
		defer chain.CloseDB()
		// -----------------------------------------------------------

		UTXOset := blockchain.UTXOSet{
			Blockchain: chain,
		}

		UTXOset.Reindex()

		count := UTXOset.CountTransactions()

		// -----------------------------------------------------------
		response := map[string]int{"txCount": count}
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.Write(jsonResponse)

	default:
		http.Error(w, "ERROR: Invalid HTTP Method", http.StatusBadRequest)
	}
}

// ------------------------------------------------------------------

func NewBlockchainServer(port uint16) *BlockchainServer {
	return &BlockchainServer{
		port: port,
	}
}

func (bcs *BlockchainServer) startNetworkServer() {
	bcs.LoadBlockchain()
	chain, _ := bcs.GetBlockchain()
	network.StartServer(chain, bcs.port, MINER_ADDRESS)
}

func (bcs *BlockchainServer) Run() {
	if err := bcs.LoadBlockchain(); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/printchain", bcs.PrintChain)
	http.HandleFunc("/newaccount", bcs.NewAccount)
	http.HandleFunc("/loadwallet", bcs.LoadWallet)
	http.HandleFunc("/getblock", bcs.GetBlock)
	http.HandleFunc("/utxoset", bcs.GetUTXOset)
	http.HandleFunc("/balance", bcs.GetBalance)
	http.HandleFunc("/reindex", bcs.Reindex)
	http.HandleFunc("/gettxn", bcs.GetTXN)
	http.HandleFunc("/addtxn", bcs.AddTXN)

	go bcs.startNetworkServer()

	hostURL := fmt.Sprintf("0.0.0.0:%d", bcs.port)
	fmt.Println("Blockchain HTTP Server is live @:", hostURL)
	log.Fatal(http.ListenAndServe(hostURL, nil))
}
