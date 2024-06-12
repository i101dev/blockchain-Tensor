package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/i101dev/blockchain-Tensor/blockchain"
	"github.com/i101dev/blockchain-Tensor/types"
	"github.com/i101dev/blockchain-Tensor/wallet"
)

var (
	cache map[string]*blockchain.Blockchain = make(map[string]*blockchain.Blockchain)

	CHAIN_ID       = "blockchain"
	ORIGIN_ADDRESS = "1MDghwANCEEUnbiCsdGPUTM9AVmfFr8auK"
)

type BlockchainServer struct {
	port uint16
}

func NewBlockchainServer(port uint16) *BlockchainServer {
	return &BlockchainServer{
		port: port,
	}
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

		var txn types.NewTxnReq
		decoder := json.NewDecoder(req.Body)

		// ----------------------------------------------------------
		err := decoder.Decode(&txn)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// ----------------------------------------------------------
		bc, err := bcs.GetBlockchain()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		blockchain.OpenDB(bc)
		defer bc.CloseDB()

		// ----------------------------------------------------------
		wallet, _ := wallet.CreateWallets()
		newTxn := blockchain.NewTransaction(txn.From, txn.To, txn.Amount, bc, wallet)
		err = bc.AddBlock([]*blockchain.Transaction{newTxn})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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

		// -----------------------------------------------------------
		walletDat, _ := wallet.CreateWallets()
		account := walletDat.GetAccount(address)

		utxoset := chain.FindUTXO(wallet.PublicKeyHash(account.PublicKey))

		balance := 0
		for _, output := range utxoset {
			balance += output.Value
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

	http.HandleFunc("/gettxn", bcs.GetTXN)
	http.HandleFunc("/addtxn", bcs.AddTXN)

	hostURL := "0.0.0.0:" + strconv.Itoa(int(bcs.port))

	fmt.Println("Blockchain Server is live @:", hostURL)
	log.Fatal(http.ListenAndServe(hostURL, nil))
}
