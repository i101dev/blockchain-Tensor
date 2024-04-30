package main

import (
	"os"

	"github.com/i101dev/golang-blockchain/cli"
)

func main() {
	defer os.Exit(0)
	c := cli.CommandLine{}
	c.Run()

	// w := wallet.MakeWallet()
	// w.Address()
}
