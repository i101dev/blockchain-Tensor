package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"fmt"
)

type Block struct {
	Timestamp    int64
	Nonce        int
	PrevHash     []byte
	Hash         []byte
	Transactions []*Transaction
}

func Genesis(coinbase *Transaction) (*Block, error) {
	return CreateBlock([]*Transaction{coinbase}, []byte{})
}

func CreateBlock(txs []*Transaction, prevHash []byte) (*Block, error) {

	block := &Block{
		PrevHash:     prevHash,
		Hash:         []byte{},
		Transactions: txs,
		Nonce:        0,
	}

	pow := NewProof(block)
	nonce, hash, err := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block, err
}

func Deserialize(data []byte) (*Block, error) {

	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(data))

	if err := decoder.Decode(&block); err != nil {
		return nil, fmt.Errorf("Failed to decode and deserialize bytes in to Block")
	}

	return &block, nil
}

func (b *Block) Serialize() ([]byte, error) {

	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)

	if err := encoder.Encode(b); err != nil {
		return nil, fmt.Errorf("failed to encode block to bytes")
	}

	return res.Bytes(), nil
}

func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}

	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}

func (b *Block) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Timestamp int64  `json:"timestamp"`
		Nonce     int    `json:"nonce"`
		Hash      string `json:"hash"`
		PrevHash  string `json:"previous_hash"`
		// Data      string `json:"data"`
	}{
		Timestamp: b.Timestamp,
		Nonce:     b.Nonce,
		Hash:      fmt.Sprintf("%x", b.Hash),
		PrevHash:  fmt.Sprintf("%x", b.PrevHash),
		// Data:      string(b.Data),
	})
}
