package blockchain

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type Block struct {
	Timestamp    int64
	Height       int
	Nonce        int
	PrevHash     []byte
	Hash         []byte
	Transactions []*Transaction
}

func (b *Block) Print() {

	fmt.Printf("\n> Hash:		%s", hex.EncodeToString(b.Hash))
	fmt.Printf("\n> PrevHash:	%x", b.PrevHash)
	fmt.Printf("\n\n> Nonce:	%d", b.Nonce)
	fmt.Printf("\n> Timestamp:	%d", b.Timestamp)

	pow := NewProof(b)
	isValid, _ := pow.Validate()

	fmt.Printf("\n> Valid Proof: 	%s", strconv.FormatBool(isValid))
	fmt.Println("\n\n### Transactions:")
	for _, t := range b.Transactions {
		t.Print()
	}
	fmt.Println()
}

func (b *Block) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Timestamp    int64          `json:"timestamp"`
		Nonce        int            `json:"nonce"`
		PrevHash     string         `json:"prev_hash"`
		Hash         string         `json:"hash"`
		Transactions []*Transaction `json:"transactions"`
	}{
		Timestamp:    b.Timestamp,
		Nonce:        b.Nonce,
		PrevHash:     hex.EncodeToString(b.PrevHash),
		Hash:         hex.EncodeToString(b.Hash),
		Transactions: b.Transactions,
	})
}

func (b *Block) Serialize() []byte {

	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)

	if err := encoder.Encode(b); err != nil {
		return nil
	}

	return res.Bytes()
}

func (b *Block) HashTransactions() []byte {

	var txHashes [][]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.Serialize())
	}

	tree := NewMerkleTree(txHashes)

	return tree.RootNode.Data
}

func Genesis(coinbase *Transaction) (*Block, error) {
	return CreateBlock([]*Transaction{coinbase}, []byte{}, 0)
}

func CreateBlock(txs []*Transaction, prevHash []byte, height int) (*Block, error) {

	block := &Block{
		Timestamp:    time.Now().UnixNano(),
		Height:       height,
		Nonce:        0,
		PrevHash:     prevHash,
		Hash:         []byte{},
		Transactions: txs,
	}

	pow := NewProof(block)
	nonce, hash, err := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block, err
}

func DeserializeBlock(data []byte) (*Block, error) {

	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(data))

	if err := decoder.Decode(&block); err != nil {
		return nil, fmt.Errorf("failed to decode and deserialize bytes in to Block")
	}

	return &block, nil
}
