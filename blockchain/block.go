package blockchain

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"

	"github.com/i101dev/blockchain-Tensor/util"
)

type Block struct {
	Timestamp int64
	Nonce     int
	PrevHash  []byte
	Hash      []byte
	Data      []byte
}

func Genesis() *Block {
	return CreateBlock("Genesis", []byte{})
}

func CreateBlock(data string, prevHash []byte) *Block {

	block := &Block{
		PrevHash: prevHash,
		Hash:     []byte{},
		Data:     []byte(data),
		Nonce:    0,
	}

	pow := NewProof(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func Deserialize(data []byte) (*Block, error) {

	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(data))

	if err := decoder.Decode(&block); err != nil {
		return nil, fmt.Errorf("Failed to decode and deserialize bytes in to Block")
	}

	return &block, nil
}

func (b *Block) Serialize() []byte {

	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)

	if err := encoder.Encode(b); err != nil {
		util.Handle(err, "Serialize")
	}

	return res.Bytes()
}

func (b *Block) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Timestamp int64  `json:"timestamp"`
		Nonce     int    `json:"nonce"`
		Hash      string `json:"hash"`
		PrevHash  string `json:"previous_hash"`
		Data      string `json:"data"`
	}{
		Timestamp: b.Timestamp,
		Nonce:     b.Nonce,
		Hash:      fmt.Sprintf("%x", b.Hash),
		PrevHash:  fmt.Sprintf("%x", b.PrevHash),
		Data:      string(b.Data),
	})
}
