package blockchain

import (
	"bytes"
	"crypto/sha256"
)

type Block struct {
	PrevHash []byte
	ThisHash []byte
	Data     []byte
}

func (b *Block) DeriveHash() {
	info := bytes.Join([][]byte{b.Data, b.PrevHash}, []byte{})
	hash := sha256.Sum256(info)
	b.ThisHash = hash[:]
}

func CreateBlock(data string, prevHash []byte) *Block {

	block := &Block{
		PrevHash: prevHash,
		ThisHash: []byte{},
		Data:     []byte(data),
	}

	block.DeriveHash()

	return block
}
