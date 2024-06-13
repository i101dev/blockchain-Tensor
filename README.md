## Proof of Work & UTXO Blockchain with web API

Port 5000 by default:

```
cd blockchain_server && go run .
```

Or set port manually

```
cd blockchain_server && go run . -port <PORT>
```

## API Routes

### GET /printchain

-   **Description**: Returns all blocks in the blockchain.
-   **Response**: JSON array of all blocks.

### GET /newaccount

-   **Description**: Creates a new account and adds it to the wallet.
-   **Response**: JSON representation of the new account.

### GET /loadwallet

-   **Description**: Loads and prints the wallet information.
-   **Response**: JSON representation of the wallet information.

### GET /getblock

-   **Description**: Retrieves a block by its hash.
-   **Query Parameters**:
    -   `hash`: The hash of the block to retrieve.
-   **Response**: JSON representation of the block.

### GET /gettxn

-   **Description**: Retrieves a transaction by its ID.
-   **Query Parameters**:
    -   `id`: The ID of the transaction to retrieve.
-   **Response**: JSON representation of the transaction.

### POST /addtxn

-   **Description**: Adds a new transaction to the blockchain.
-   **Request Body**: JSON object containing `from`, `to`, and `amount` fields.
-   **Response**: JSON representation of the added transaction.

### GET /utxoset

-   **Description**: Retrieves the unspent transaction outputs (UTXOs) for a given address.
-   **Query Parameters**:
    -   `address`: The address to query UTXOs for.
-   **Response**: JSON array of UTXOs.

### GET /balance

-   **Description**: Retrieves the balance for a given address.
-   **Query Parameters**:
    -   `address`: The address to query the balance for.
-   **Response**: JSON object with the balance.

### GET /reindex

-   **Description**: Reindexes the UTXO set.
-   **Response**: JSON object with the count of transactions in the UTXO set.

---

### Credit goes to Tensor for laying the foundation:

https://github.com/tensor-programming/golang-blockchain
