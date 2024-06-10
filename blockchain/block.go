package blockchain

type Block struct {
	PrevHash []byte
	ThisHash []byte
	Data     []byte
	Nonce    int
}

func CreateBlock(data string, prevHash []byte) *Block {

	block := &Block{
		PrevHash: prevHash,
		ThisHash: []byte{},
		Data:     []byte(data),
		Nonce:    0,
	}

	pow := NewProof(block)
	nonce, hash := pow.Run()

	block.ThisHash = hash[:]
	block.Nonce = nonce

	return block
}
