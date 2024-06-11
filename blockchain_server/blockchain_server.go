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
)

var (
	cache map[string]*blockchain.Blockchain = make(map[string]*blockchain.Blockchain)

	CHAIN_ID       = "blockchain"
	ORIGIN_ADDRESS = "OriginAddress"
)

type BlockchainServer struct {
	port uint16
}

func NewBlockchainServer(port uint16) *BlockchainServer {
	return &BlockchainServer{
		port: port,
	}
}

func (bcs *BlockchainServer) InitBlockchain() error {

	bc, err := blockchain.LoadBlockchain(ORIGIN_ADDRESS, bcs.port)

	if err != nil {
		return fmt.Errorf("failed to initialize new chain")
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

		db := blockchain.OpenDB(bc)
		defer bc.CloseDB()

		// ----------------------------------------------------------
		iterator := &blockchain.BlockchainIterator{
			CurrentHash: bc.LastHash,
			Database:    db,
			Chain:       bc,
		}

		// ----------------------------------------------------------
		var allBlocks []byte
		allBlocks = append(allBlocks, '[') // Start of JSON array

		it := 0
		for {
			block, err := iterator.IterateNext()
			if err != nil {
				break
			}

			m, err := block.MarshalJSON()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if it > 0 {
				allBlocks = append(allBlocks, ',') // Add comma between JSON objects
			}

			allBlocks = append(allBlocks, m...)

			it++

			if len(block.PrevHash) == 0 {
				break
			}
		}

		allBlocks = append(allBlocks, ']') // End of JSON array

		// ----------------------------------------------------------

		w.Header().Add("Content-Type", "application/json")
		w.Write(allBlocks)

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
		newTxn := blockchain.NewTransaction(txn.From, txn.To, txn.Amount, bc)
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
		var UTXO []byte
		for _, output := range utxoset {

			o, err := output.MarshalJSON()

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			UTXO = append(UTXO, o...)
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
		utxoset := chain.FindUTXO(address)

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

	if err := bcs.InitBlockchain(); err != nil {
		log.Fatal(err)
	}

	// http.HandleFunc("/initchain", bcs.InitChain)
	http.HandleFunc("/printchain", bcs.PrintChain)
	http.HandleFunc("/getblock", bcs.GetBlock)
	http.HandleFunc("/utxoset", bcs.GetUTXOset)
	http.HandleFunc("/balance", bcs.GetBalance)

	http.HandleFunc("/addtxn", bcs.AddTXN)

	hostURL := "0.0.0.0:" + strconv.Itoa(int(bcs.port))

	fmt.Println("Blockchain Server is live @:", hostURL)
	log.Fatal(http.ListenAndServe(hostURL, nil))
}
