package blockchain

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/i101dev/blockchain-Tensor/util"
	"github.com/i101dev/blockchain-Tensor/wallet"
)

// ---------------------------------------------------------------------
// type TxInput struct {
// 	ID  []byte // the respective transaction the output came from
// 	Out int    // the index within the respective transaction []TxOutput
// 	Sig string // provides the data used in the outputs `PubKey`
// }

type TxInput struct {
	Out       int
	ID        []byte
	Signature []byte
	PubKey    []byte
}

func (in *TxInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := wallet.PublicKeyHash(in.PubKey)
	return bytes.Equal(lockingHash, pubKeyHash)
}

func (in *TxInput) Print() {
	fmt.Println("    **")
	fmt.Printf("    | ID: %x\n", in.ID)
	fmt.Printf("    | Out: %d\n", in.Out)
	fmt.Printf("    | Signature: %s\n", in.Signature)
}

func (in *TxInput) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID        string `json:"id"`
		Out       int    `json:"out"`
		Signature string `json:"signature"`
		PubKey    string `json:"pubkey"`
	}{
		ID:        hex.EncodeToString(in.ID),
		Out:       in.Out,
		Signature: hex.EncodeToString(in.Signature),
		PubKey:    hex.EncodeToString(in.PubKey),
	})
}

func (in *TxInput) UnmarshalJSON(data []byte) error {
	aux := struct {
		ID        string `json:"id"`
		Out       int    `json:"out"`
		Signature string `json:"signature"`
		PubKey    string `json:"pubkey"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	id, err := hex.DecodeString(aux.ID)
	if err != nil {
		return err
	}

	signature, err := hex.DecodeString(aux.Signature)
	if err != nil {
		return err
	}

	pubKey, err := hex.DecodeString(aux.PubKey)
	if err != nil {
		return err
	}

	in.ID = id
	in.Out = aux.Out
	in.Signature = signature
	in.PubKey = pubKey

	return nil
}

// ---------------------------------------------------------------------
type TxOutput struct {
	Value      int
	PubKeyHash []byte
}

func (out *TxOutput) Print() {
	fmt.Println("    **")
	fmt.Printf("    | Value: %d\n", out.Value)
	fmt.Printf("    | PubKey: %s\n", out.PubKeyHash)
}

func (out *TxOutput) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Value      int    `json:"value"`
		PubKeyHash string `json:"pubkey_hash"`
	}{
		Value:      out.Value,
		PubKeyHash: hex.EncodeToString(out.PubKeyHash),
	})
}

func (out *TxOutput) UnmarshalJSON(data []byte) error {
	aux := struct {
		Value      int    `json:"value"`
		PubKeyHash string `json:"pubkey_hash"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	out.Value = aux.Value
	var err error
	out.PubKeyHash, err = hex.DecodeString(aux.PubKeyHash)
	if err != nil {
		return err
	}

	return nil
}

func (out *TxOutput) Lock(address []byte) {
	pubKeyHash := util.Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

func NewTXOutput(value int, address string) *TxOutput {

	txo := &TxOutput{
		Value:      value,
		PubKeyHash: nil,
	}

	txo.Lock([]byte(address))

	return txo
}
