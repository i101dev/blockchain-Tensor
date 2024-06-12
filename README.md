# Boardgame Blockchain

-   `go run main.go` to run the app
-   `go build main.go` to build an executable file

This is a proprietary blockchain that I've been meditating on, with the intention to integrate a digital boardgame I already built using the MERN stack.

Because this is for a game that I already built, there is no penalty for failure. If something with the chain goes wrong, no problem. Give the player with the highest score a prize and shut it down. Find and fix whatever bug caused the issue, and start a new game.

This strategy has the added benefit of neutralizing incumbant players and gives everyone a fresh and fair start.

Blockchain can gaurantee integrity of the game and neutralize cheaters immediately.

# Concept

-   To audit the chain is a right

-   To run a node is a privilidge

-   Node operators can block known cheaters

-   Perfect decentralization not required for this particular use case

-   The sum of a players actions - transactions - during their turn is the `block.Data`

-   The block is recorded when the active player ends their turn - player sends TX free of charge

---

Credit goes to Tensor for laying the foundation:

https://github.com/tensor-programming/golang-blockchain
