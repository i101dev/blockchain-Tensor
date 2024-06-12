package blockchain

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
)

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
