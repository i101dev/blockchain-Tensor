package blockchain

import (
	"bytes"
	"encoding/gob"

	"github.com/i101dev/blockchain-Tensor/wallet2"
)

type TxOutput struct {
	Value      int
	PubKeyHash []byte
}

type TxOutputs struct {
	Outputs []TxOutput
}

type TxInput struct {
	HashID    []byte
	Out       int
	Signature []byte
	PubKey    []byte
}

func NewTXOutput(value int, address string) *TxOutput {
	txo := &TxOutput{value, nil}
	txo.Lock([]byte(address))
	return txo
}

func DeserializeOutputs(data []byte) TxOutputs {
	var outputs TxOutputs
	decode := gob.NewDecoder(bytes.NewReader(data))
	err := decode.Decode(&outputs)
	Handle(err)
	return outputs
}

func (input *TxInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := wallet2.PublicKeyHash(input.PubKey)
	return bytes.Equal(lockingHash, pubKeyHash)
}

func (output *TxOutput) Lock(address []byte) {
	pubKeyHash := wallet2.Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	output.PubKeyHash = pubKeyHash
}

func (output *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Equal(output.PubKeyHash, pubKeyHash)
}

func (outputs TxOutputs) Serialize() []byte {
	var buffer bytes.Buffer
	encode := gob.NewEncoder(&buffer)
	err := encode.Encode(outputs)
	Handle(err)
	return buffer.Bytes()
}
