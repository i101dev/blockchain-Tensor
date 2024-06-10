package main

import (
	"log"

	// "github.com/i101dev/blockchain-Tensor/cli"

	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func main() {
	// fmt.Println("Yo dawg")
	// chain := blockchain.InitBlockChain(os.Getenv("PORT"))

	// chain.AddBlock("Block 1")
	// chain.AddBlock("Block 2")
	// chain.AddBlock("Block 3")

	// for i, block := range chain.Blocks {
	// 	fmt.Printf("\nBlock %d %s", i, strings.Repeat("-", 80))
	// 	fmt.Printf("\nPrevHash: %+x", block.PrevHash)
	// 	fmt.Printf("\nHash: %x", block.ThisHash)
	// 	fmt.Printf("\nData: %s\n", block.Data)

	// 	pow := blockchain.NewProof(block)
	// 	isValid := pow.Validate()
	// 	fmt.Printf("PoW: %s\n", strconv.FormatBool(isValid))
	// 	fmt.Println()
	// }
}
