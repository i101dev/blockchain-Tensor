package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

type Transaction struct {
	HashID  []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

type TxOutput struct {
	Value  int
	PubKey string
}

type TxInput struct {
	HashID []byte
	Out    int
	Sig    string
}

func (tx *Transaction) SetID() {

	var encoded bytes.Buffer
	var hash [32]byte

	encode := gob.NewEncoder(&encoded)
	err := encode.Encode(tx)

	Handle(err)

	hash = sha256.Sum256(encoded.Bytes())

	tx.HashID = hash[:]
}

func (tx *Transaction) IsCoinbase() bool {

	chk1 := len(tx.Inputs) == 1
	chk2 := len(tx.Inputs[0].HashID) == 0
	chk3 := tx.Inputs[0].Out == -1

	return chk1 && chk2 && chk3
}

func (input *TxInput) CanUnlock(data string) bool {
	return input.Sig == data
}

func (output *TxOutput) CanBeUnlocked(data string) bool {
	return output.PubKey == data
}

func CoinbaseTx(to string, data string) *Transaction {

	if data == "" {
		data = fmt.Sprintf("Coins to %s", to)
	}

	txin := TxInput{[]byte{}, -1, data}
	txout := TxOutput{100, to}

	tx := Transaction{nil, []TxInput{txin}, []TxOutput{txout}}
	tx.SetID()

	return &tx
}

func NewTransaction(from string, to string, amount int, chain *Blockchain) *Transaction {

	var inputs []TxInput
	var outputs []TxOutput

	acc, validOutputs := chain.FindSpendableOutputs(from, amount)

	if acc < amount {
		log.Panic("Error: insufficient funds")
	}

	for txId, outs := range validOutputs {
		txId, err := hex.DecodeString(txId)
		Handle(err)

		for _, out := range outs {
			newInput := TxInput{txId, out, from}
			inputs = append(inputs, newInput)
		}
	}

	outputs = append(outputs, TxOutput{amount, to})

	if acc > amount {
		outputs = append(outputs, TxOutput{acc - amount, from})
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx
}
