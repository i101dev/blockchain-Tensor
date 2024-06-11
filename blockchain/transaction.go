package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"

	"github.com/i101dev/blockchain-Tensor/util"
)

type Transaction struct {
	ID      []byte // the hash of the transaction
	Inputs  []TxInput
	Outputs []TxOutput
}

type TxInput struct {
	ID  []byte // the respective transaction the output came from
	Out int    // the index within the respective transaction []TxOutput
	Sig string // provides the data used in the outputs `PubKey`
}

type TxOutput struct {
	Value  int
	PubKey string
}

func (tx *Transaction) SetID() {

	var encoded bytes.Buffer
	var hash [32]byte

	encode := gob.NewEncoder(&encoded)
	err := encode.Encode(tx)
	util.Handle(err, "** SetID - transaction.go **")

	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

func (tx *Transaction) IsCoinbase() bool {
	lenOne := len(tx.Inputs) == 1
	idZero := len(tx.Inputs[0].ID) == 0
	outOne := tx.Inputs[0].Out == -1
	return lenOne && idZero && outOne
}

func (in *TxInput) CanUnlock(signature string) bool {
	return in.Sig == signature
}

func (out *TxOutput) CanBeUnlocked(publicKey string) bool {
	return out.PubKey == publicKey
}

func CoinbaseTX(to string, data string) *Transaction {

	if data == "" {
		data = fmt.Sprintf("Coins to %s", to)
	}

	txIn := TxInput{
		ID:  []byte{},
		Out: -1,
		Sig: data,
	}

	txOut := TxOutput{
		Value:  100,
		PubKey: to,
	}

	newTX := Transaction{
		ID:      nil,
		Inputs:  []TxInput{txIn},
		Outputs: []TxOutput{txOut},
	}

	newTX.SetID()

	return &newTX
}
