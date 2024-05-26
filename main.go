package main

import (
	"log"
	"os"

	"github.com/i101dev/blockchain-Tensor/cli"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func main() {
	defer os.Exit(0)
	c := cli.CommandLine{}
	c.Run()

	// log.SetFlags(0)

	// http.HandleFunc("/createblockchain", api.HandleCreateBlockchain)
	// http.HandleFunc("/getbalance", api.HandleGetBalance)
	// http.HandleFunc("/send", api.HandleSend)
	// http.HandleFunc("/startnode", api.HandleStartNode)
	// http.HandleFunc("/listaddresses", api.HandleListAddresses)
	// http.HandleFunc("/createwallet", api.HandleCreateWallet)
	// http.HandleFunc("/reindexutxo", api.HandleReindexUTXO)
	// http.HandleFunc("/printchain", api.HandlePrintChain)
	// http.HandleFunc("/listnodes", api.HandleListNodes)

	// nodeID := os.Getenv("NODE_ID")
	// port := os.Getenv("PORT")

	// if nodeID == "" {
	// 	log.Fatal("*** >>> NODE_ID env is not set <<< ***")
	// 	runtime.Goexit()
	// }

	// if port == "" {
	// 	log.Fatal("*** >>> PORT env is not set <<< ***")
	// 	runtime.Goexit()
	// }

	// log.Printf("Server started on Port: %s", port)
	// log.Fatal(http.ListenAndServe(":"+port, nil))

	// wallet2.Test()
}
