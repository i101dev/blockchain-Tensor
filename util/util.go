package util

import (
	"fmt"
	"log"
)

func Handle(err error, funcName string) {
	if err != nil {
		fmt.Printf("\n*** >>> ERROR: [%s] <<< *** \n", funcName)
		log.Panic(err)
	}
}
