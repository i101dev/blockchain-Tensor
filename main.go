package main

import (
	"fmt"
	"strconv"

	"github.com/i101dev/golang-blockchain/blockchain"
)

func main() {
	// fmt.Println("Online and working fine")
	chain := blockchain.InitBlockChain()

	chain.AddBlock("First block")
	chain.AddBlock("Second block")
	chain.AddBlock("Third block")

	for i := 0; i < len(chain.Blocks); i++ {
		fmt.Printf("\nPrevious hash: %x", chain.Blocks[i].PrevHash)
		fmt.Printf("\nBlock data: %s", chain.Blocks[i].Data)
		fmt.Printf("\nBlock hash: %x\n", chain.Blocks[i].Hash)

		pow := blockchain.NewProof(chain.Blocks[i])

		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
	}
}
