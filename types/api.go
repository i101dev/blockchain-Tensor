package types

type AddBlockReq struct {
	Data *string `json:"data"`
}

type NewTxnReq struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Amount  int    `json:"amount"`
	MineNow bool   `json:"minenow"`
}
