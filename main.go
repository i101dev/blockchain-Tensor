package main

import (
	"log"
	"net/http"
	"os"
	"runtime"

	"github.com/i101dev/blockchain-Tensor/api"
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

	http.HandleFunc("/createblockchain", api.HandleCreateBlockchain)
	http.HandleFunc("/getbalance", api.HandleGetBalance)
	http.HandleFunc("/send", api.HandleSend)
	http.HandleFunc("/startnode", api.HandleStartNode)
	http.HandleFunc("/listaddresses", api.HandleListAddresses)
	http.HandleFunc("/createwallet", api.HandleCreateWallet)
	http.HandleFunc("/reindexutxo", api.HandleReindexUTXO)
	http.HandleFunc("/printchain", api.HandlePrintChain)

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
