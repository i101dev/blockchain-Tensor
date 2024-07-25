package main

import (
	"flag"
	"log"
	"os"
)

func init() {
	log.SetPrefix("Blockchain API: ")
}

func main() {

	defer os.Exit(0)

	port := flag.Uint("port", 5000, "TCP Port Number for API Server")
	flag.Parse()

	app := NewBlockchainServer(uint16(*port))

	app.Run()
}
