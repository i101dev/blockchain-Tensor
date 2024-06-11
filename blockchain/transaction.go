package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
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

func NewTransaction(from string, to string, amount int, chain *Blockchain) (*Transaction, error) {

	var inputs []TxInput
	var outputs []TxOutput

	acc, validOutputs := chain.FindSpendableOutputs(from, amount)

	if acc < amount {
		return nil, fmt.Errorf("ERROR: insufficient funds")
	}

	for outID, outs := range validOutputs {
		txID, err := hex.DecodeString(outID)
		if err != nil {
			return nil, err
		}

		for _, out := range outs {
			inp := TxInput{
				ID:  txID,
				Out: out,
				Sig: from,
			}
			inputs = append(inputs, inp)
		}
	}

	// ---------------------------------------------------
	newOutput1 := TxOutput{
		Value:  amount,
		PubKey: to,
	}
	outputs = append(outputs, newOutput1)

	// ---------------------------------------------------
	if acc > amount {
		newOutput2 := TxOutput{
			Value:  acc - amount,
			PubKey: from,
		}
		outputs = append(outputs, newOutput2)
	}

	// ---------------------------------------------------
	newTx := &Transaction{
		Inputs:  inputs,
		Outputs: outputs,
	}

	newTx.SetID()

	return newTx, nil
}

func DummyTransaction(from string, to string, amount int, chain *Blockchain) (*Transaction, error) {

	var inputs []TxInput
	var outputs []TxOutput

	originTxHash, _ := hex.DecodeString("ad0f92fb8e1489c59cbc7833ca2e19581fd72b6b856e4b0e37d0694a8dc86930")

	// ---------------------------------------------------
	inputs = append(inputs, TxInput{
		ID:  originTxHash,
		Out: 0,
		Sig: from,
	})
	// ---------------------------------------------------
	outputs = append(outputs, TxOutput{
		Value:  amount,
		PubKey: to,
	})

	// ---------------------------------------------------
	newTx := &Transaction{
		Inputs:  inputs,
		Outputs: outputs,
	}

	newTx.SetID()

	return newTx, nil
}

// ---------------------------------------------------------------------
type TxInput struct {
	ID  []byte // the respective transaction the output came from
	Out int    // the index within the respective transaction []TxOutput
	Sig string // provides the data used in the outputs `PubKey`
}

func (in *TxInput) CanUnlock(signature string) bool {
	return in.Sig == signature
}

func (in *TxInput) Print() {
	fmt.Println("    **")
	fmt.Printf("    | TxInput ID: %x\n", in.ID)
	fmt.Printf("    | Out: %d\n", in.Out)
	fmt.Printf("    | Signature: %s\n", in.Sig)
}

func (in *TxInput) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID  string `json:"id"`
		Out int    `json:"out"`
		Sig string `json:"sig"`
	}{
		ID:  hex.EncodeToString(in.ID),
		Out: in.Out,
		Sig: in.Sig,
	})
}

func (in *TxInput) UnmarshalJSON(data []byte) error {
	aux := struct {
		ID  string `json:"id"`
		Out int    `json:"out"`
		Sig string `json:"sig"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	id, err := hex.DecodeString(aux.ID)
	if err != nil {
		return err
	}

	in.ID = id
	in.Out = aux.Out
	in.Sig = aux.Sig

	return nil
}

// ---------------------------------------------------------------------
type TxOutput struct {
	Value  int
	PubKey string
}

func (out *TxOutput) Print() {
	fmt.Println("    **")
	fmt.Printf("    | Value: %d\n", out.Value)
	fmt.Printf("    | PubKey: %s\n", out.PubKey)
}

func (out *TxOutput) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Value  int    `json:"value"`
		PubKey string `json:"pubkey"`
	}{
		Value:  out.Value,
		PubKey: out.PubKey,
	})
}

func (out *TxOutput) UnmarshalJSON(data []byte) error {
	aux := struct {
		Value  int    `json:"value"`
		PubKey string `json:"pubkey"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	out.Value = aux.Value
	out.PubKey = aux.PubKey

	return nil
}

func (out *TxOutput) CanBeUnlocked(publicKey string) bool {
	return out.PubKey == publicKey
}
