package blockchain

type TxOutput struct {
	Value  int
	PubKey string
}

type TxInput struct {
	HashID []byte
	Out    int
	Sig    string
}

func (input *TxInput) CanUnlock(data string) bool {
	return input.Sig == data
}

func (output *TxOutput) CanBeUnlocked(data string) bool {
	return output.PubKey == data
}
