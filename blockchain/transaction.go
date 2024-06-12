package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/i101dev/blockchain-Tensor/util"
)

type Transaction struct {
	ID      []byte // the hash of the transaction
	Inputs  []TxInput
	Outputs []TxOutput
}

func (t *Transaction) Print() {
	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("> ID: %x", t.ID)
	fmt.Println("\n> Inputs:")
	for _, input := range t.Inputs {
		input.Print()
	}
	fmt.Println("\n> Outputs:")
	for _, output := range t.Outputs {
		output.Print()
	}
}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID      string     `json:"id"`
		Inputs  []TxInput  `json:"inputs"`
		Outputs []TxOutput `json:"outputs"`
	}{
		ID:      hex.EncodeToString(t.ID),
		Inputs:  t.Inputs,
		Outputs: t.Outputs,
	})
}

func (t *Transaction) UnmarshalJSON(data []byte) error {
	aux := struct {
		ID      string     `json:"id"`
		Inputs  []TxInput  `json:"inputs"`
		Outputs []TxOutput `json:"outputs"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	id, err := hex.DecodeString(aux.ID)
	if err != nil {
		return err
	}

	t.ID = id
	t.Inputs = aux.Inputs
	t.Outputs = aux.Outputs

	return nil
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

func NewTransaction(from, to string, amount int, chain *Blockchain) *Transaction {

	// blockchain.OpenDB(chain)
	// defer chain.CloseDB()

	var inputs []TxInput
	var outputs []TxOutput

	acc, validOutputs := chain.FindSpendableOutputs(from, amount)

	if acc < amount {
		log.Panic("Error: not enough funds")
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		util.Handle(err, "NewTransaction 1")

		for _, out := range outs {
			input := TxInput{txID, out, from}
			inputs = append(inputs, input)
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
