package wallet

import (
	"fmt"

	"github.com/mr-tron/base58"
)

func Base58Encode(input []byte) []byte {
	encode := base58.Encode(input)
	return []byte(encode)
}

func Base58Decode(input []byte) []byte {
	decode, err := base58.Decode(string(input[:]))
	Handle(err, "Base58Decode")
	return decode
}

func Handle(err error, functionName string) {
	if err != nil {
		fmt.Printf(`\n*** >>> ERROR @ [%s] - %+v`, functionName, err)
	}
}
