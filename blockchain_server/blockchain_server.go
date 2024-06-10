package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/i101dev/blockchain-Tensor/blockchain"
	"github.com/i101dev/blockchain-Tensor/types"
)

var (
	cache map[string]*blockchain.Blockchain = make(map[string]*blockchain.Blockchain)

	CHAIN_ID = "blockchain"
)

type BlockchainServer struct {
	port uint16
}

func NewBlockchainServer(port uint16) *BlockchainServer {
	return &BlockchainServer{
		port: port,
	}
}

func (bcs *BlockchainServer) GetBlockchain(id string) *blockchain.Blockchain {

	bc, ok := cache[id]

	if !ok {
		// minerWallet := wallet.NewWallet()
		// log.Printf("\nAddress: %s", minerWallet.BlockchainAddress())
		bc = blockchain.InitBlockchain("MiningAddress", bcs.port)
		cache[id] = bc
	}

	return bc
}

func (bcs *BlockchainServer) GetChainData(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:

		w.Header().Add("Content-Type", "application/json")

		bc := bcs.GetBlockchain(CHAIN_ID)
		db := blockchain.OpenDB(bc)

		fmt.Printf("\n*** >>> GetChainData %s", strings.Repeat("-", 70))
		fmt.Printf("\nLast Hash %+x", bc.LastHash)

		iterator := &blockchain.BlockchainIterator{
			CurrentHash: bc.LastHash,
			Database:    db,
			Chain:       bc,
		}

		it := 0
		for {

			block := iterator.IterateNext()

			fmt.Printf("\n%s", strings.Repeat("=", 80))
			fmt.Printf("\n*** Block %d %s\n", it, strings.Repeat("-", 67))
			fmt.Printf("\nLast: %x", block.PrevHash)
			fmt.Printf("\nHash: %x", block.Hash)
			fmt.Printf("\nData: %s\n", block.Data)

			pow := blockchain.NewProof(block)
			isValid, _ := pow.Validate()
			fmt.Printf("PoW: %s\n", strconv.FormatBool(isValid))
			fmt.Println()
			if len(block.PrevHash) == 0 {
				break
			}

			it++
		}

		bc.CloseDB()

	default:
		io.WriteString(w, "ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) BlockByHash(w http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case http.MethodGet:

		hash := req.URL.Query().Get("hash")
		hashBytes, err := hex.DecodeString(hash)

		if err != nil {
			io.WriteString(w, string(err.Error()[:]))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// ----------------------------------------------------------
		bc := bcs.GetBlockchain(CHAIN_ID)
		db := blockchain.OpenDB(bc)

		block, err := bc.GetBlockByHash(db, hashBytes)
		bc.CloseDB()

		if err != nil {
			io.WriteString(w, string(err.Error()[:]))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// ----------------------------------------------------------
		m, err := block.MarshalJSON()
		if err != nil {
			io.WriteString(w, string(err.Error()[:]))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application-json")
		io.WriteString(w, string(m[:]))

	default:
		io.WriteString(w, "ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) AddBlock(w http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case http.MethodPost:

		var data types.AddBlockReq

		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&data)

		if err != nil {
			io.WriteString(w, string(err.Error()[:]))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		bcs.GetBlockchain(CHAIN_ID).AddBlock(&data)

	default:
		io.WriteString(w, "ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) Run() {

	// bcs.GetBlockchain().Run()

	http.HandleFunc("/chaindata", bcs.GetChainData)
	http.HandleFunc("/blockbyhash", bcs.BlockByHash)
	http.HandleFunc("/addblock", bcs.AddBlock)

	hostURL := "0.0.0.0:" + strconv.Itoa(int(bcs.port))

	fmt.Println("Blockchain Server is live @:", hostURL)
	log.Fatal(http.ListenAndServe(hostURL, nil))
}
