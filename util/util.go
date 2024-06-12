package util

import (
	"fmt"
	"log"

	"github.com/mr-tron/base58"
)

func Handle(err error, funcName string) {
	if err != nil {
		fmt.Printf("\n*** >>> ERROR: [%s] <<< *** \n", funcName)
		log.Panic(err)
	}
}

func Base58Encode(input []byte) []byte {
	encode := base58.Encode(input)
	return []byte(encode)
}

func Base58Decode(input []byte) []byte {
	decode, err := base58.Decode(string(input[:]))
	Handle(err, "Base58Decode")
	return decode
}
