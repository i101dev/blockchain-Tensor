# Golang Blockchain

-   `go run main.go` to run the app
-   `go build main.go` to build an executable file.

This is a proprietary blockchain that I've been meditating on, with the intention to integrate a digital boardgame I already built using the MERN stack.

Because this is for a game that I already built, there is no penalty for failure. If something with the chain goes wrong, no problem. Give the player with the highest score a prize and shut it down. Find and fix whatever bug caused the issue, and start a new game.

This strategy has the added benefit of neutralizing incumbant players and gives everyone a fresh and fair start.

Blockchain tech can gaurantee integrity of the game and neutralize cheaters immediately.

# Concept

-   Perfect decentralization not required for this particular use case.

-   Is designed to ensure honest gameplay

-   Not designed to enrich speculators

-   To audit the chain is a right

-   To run a node is a privilidge

-   A player's turn constitutes a block.

-   The block is recorded when the active player ends their turn.

-   The sum of a players actions - transactions - during their turn is the `block.Data`.
