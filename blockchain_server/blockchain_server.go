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

var cache map[string]*blockchain.Blockchain = make(map[string]*blockchain.Blockchain)

type BlockchainServer struct {
	port uint16
}

func NewBlockchainServer(port uint16) *BlockchainServer {
	return &BlockchainServer{
		port: port,
	}
}

func (bcs *BlockchainServer) Port() uint16 {
	return bcs.port
}

func (bcs *BlockchainServer) GetBlockchain() *blockchain.Blockchain {

	ID := "blockchain"

	bc, ok := cache[ID]

	if !ok {
		fmt.Println("\nInitializing new chain")
		// minerWallet := wallet.NewWallet()
		bc = blockchain.InitBlockchain("MiningAddress", bcs.Port())
		cache[ID] = bc
		// log.Printf("\nAddress: %s", minerWallet.BlockchainAddress())
	} else {
		// fmt.Println("\nPulling existing chain")
	}

	return bc
}

func (bcs *BlockchainServer) GetChainData(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		w.Header().Add("Content-Type", "application/json")
		bc := bcs.GetBlockchain()

		fmt.Printf("\n*** >>> GetChainData %s", strings.Repeat("-", 70))
		fmt.Printf("\n %+x", bc.LastHash)
	default:
		log.Printf("ERROR: Invalid HTTP Method")
	}

}

func (bcs *BlockchainServer) AddBlock(w http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case http.MethodPost:

		var data types.AddBlockReq

		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&data)

		if err != nil {
			log.Printf("ERROR decoding block data: %+v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		bcs.GetBlockchain().AddBlock(&data)

	default:
		log.Printf("ERROR: Invalid HTTP Method")
	}
}

func (bcs *BlockchainServer) BlockByHash(w http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case http.MethodGet:

		hash := req.URL.Query().Get("hash")
		hashBytes, err := hex.DecodeString(hash)

		if err != nil {
			fmt.Println("\n*** >>> [hex.DecodeString(hash)] - FAIL", err)
		}

		block := bcs.GetBlockchain().GetBlockByHash(hashBytes)
		m, err := block.MarshalJSON()

		if err != nil {
			fmt.Println("\n*** >>> [block.MarshalJSON()] - FAIL", err)
		}

		w.Header().Add("Content-Type", "app")
		io.WriteString(w, string(m[:]))

	default:
		log.Println("ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) Run() {

	// bcs.GetBlockchain().Run()

	http.HandleFunc("/addblock", bcs.AddBlock)
	http.HandleFunc("/blockbyhash", bcs.BlockByHash)
	http.HandleFunc("/", bcs.GetChainData)

	hostURL := "0.0.0.0:" + strconv.Itoa(int(bcs.Port()))

	fmt.Println("Blockchain Server is live @:", hostURL)
	log.Fatal(http.ListenAndServe(hostURL, nil))
}
