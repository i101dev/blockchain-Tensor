package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/i101dev/blockchain-Tensor/util"
	"github.com/i101dev/blockchain-Tensor/wallet"
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

func (t Transaction) Serialize() []byte {

	var encoded bytes.Buffer

	encoder := gob.NewEncoder(&encoded)

	if err := encoder.Encode(t); err != nil {
		util.Handle(err, "Serialize Transaction")
	}

	return encoded.Bytes()
}

func (t *Transaction) Hash() []byte {

	var hash [32]byte

	txCopy := *t
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

func (t *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if t.IsCoinbase() {
		return
	}

	for _, in := range t.Inputs {
		if prevTXs[hex.EncodeToString(in.ID)].ID == nil {
			log.Panic("\n *** >>> ERROR: previous transaction is not correct")
		}
	}

	txCopy := t.TrimmedCopy()

	for inId, in := range txCopy.Inputs {

		prevTX := prevTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inId].PubKey = prevTX.Outputs[in.Out].PubKeyHash
		txCopy.Inputs[inId].Signature = nil

		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[inId].PubKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
		util.Handle(err, "Sign Transaction")

		signature := append(r.Bytes(), s.Bytes()...)

		t.Inputs[inId].Signature = signature
	}
}

func (t *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if t.IsCoinbase() {
		return true
	}

	// Verify each of the inputs exists
	for _, in := range t.Inputs {
		if prevTXs[hex.EncodeToString(in.ID)].ID == nil {
			log.Panic("Previous transaction not correct")
		}
	}

	txCopy := t.TrimmedCopy()
	curve := elliptic.P256()

	for inId, in := range t.Inputs {

		// Same as what's in the Transaction.Sign() method
		prevTx := prevTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inId].Signature = nil
		txCopy.Inputs[inId].PubKey = prevTx.Outputs[in.Out].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[inId].PubKey = nil

		// Deconstruct the signature
		r := big.Int{}
		s := big.Int{}
		sigLen := len(in.Signature)
		r.SetBytes(in.Signature[:(sigLen / 2)])
		s.SetBytes(in.Signature[(sigLen / 2):])

		// Deconstruct the public key
		x := big.Int{}
		y := big.Int{}
		keyLen := len(in.PubKey)
		x.SetBytes(in.PubKey[:(keyLen / 2)])
		y.SetBytes(in.PubKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{
			Curve: curve,
			X:     &x,
			Y:     &y,
		}

		if !ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) {
			return false
		}
	}

	return true
}

func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	for _, in := range tx.Inputs {
		inputs = append(inputs, TxInput{in.Out, in.ID, nil, nil})
	}

	for _, out := range tx.Outputs {
		outputs = append(outputs, TxOutput{out.Value, out.PubKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}

	return txCopy
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

// func (t *Transaction) UnmarshalJSON(data []byte) error {
// 	aux := struct {
// 		ID      string     `json:"id"`
// 		Inputs  []TxInput  `json:"inputs"`
// 		Outputs []TxOutput `json:"outputs"`
// 	}{}

// 	if err := json.Unmarshal(data, &aux); err != nil {
// 		return err
// 	}

// 	id, err := hex.DecodeString(aux.ID)
// 	if err != nil {
// 		return err
// 	}

// 	t.ID = id
// 	t.Inputs = aux.Inputs
// 	t.Outputs = aux.Outputs

// 	return nil
// }

func (tx *Transaction) IsCoinbase() bool {
	lenOne := len(tx.Inputs) == 1
	idZero := len(tx.Inputs[0].ID) == 0
	outOne := tx.Inputs[0].Out == -1
	return lenOne && idZero && outOne
}

func CoinbaseTX(to string, data string) *Transaction {

	if data == "" {
		randData := make([]byte, 24)
		_, err := rand.Read(randData)
		if err != nil {
			util.Handle(err, "CoinbaseTX")
		}

		data = fmt.Sprintf("%x", randData)
	}

	txIn := TxInput{
		ID:        []byte{},
		Out:       -1,
		Signature: nil,
		PubKey:    []byte(data),
	}

	txOut := NewTXOutput(20, to)

	newTX := Transaction{
		ID:      nil,
		Inputs:  []TxInput{txIn},
		Outputs: []TxOutput{*txOut},
	}

	newTX.ID = newTX.Hash()

	return &newTX
}

func NewTransaction(from, to string, amount int, UTXO *UTXOSet, senderWallet *wallet.Wallet) *Transaction {

	// blockchain.OpenDB(chain)
	// defer chain.CloseDB()

	var inputs []TxInput
	var outputs []TxOutput

	w := senderWallet.GetAccount(from)
	pubKeyHash := wallet.PublicKeyHash(w.PublicKey)

	acc, validOutputs := UTXO.FindSpendableOutputs(pubKeyHash, amount)

	if acc < amount {
		log.Panic("Error: not enough funds")
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		util.Handle(err, "NewTransaction 1")

		for _, out := range outs {
			input := TxInput{out, txID, nil, w.PublicKey}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, *NewTXOutput(amount, to))

	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from))
	}

	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	UTXO.Blockchain.SignTransaction(&tx, w.PrivateKey)

	return &tx
}

func DeserializeTransaction(data []byte) Transaction {
	var transaction Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&transaction)
	util.Handle(err, "DeserializeTransaction")
	return transaction
}
