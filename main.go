package main

import (
	"os"

	"github.com/i101dev/golang-blockchain/cli"
)

func main() {
	defer os.Exit(0)
	c := cli.CommandLine{}
	c.Run()

	// w := wallet.MakeWallet()
	// w.Address()
}

type Person struct {
	Name string
	Age  int
}

// func main() {
// 	person := Person{Name: "John Doe", Age: 30}

// 	// Encode with Gob
// 	var buf bytes.Buffer
// 	enc := gob.NewEncoder(&buf)
// 	if err := enc.Encode(person); err != nil {
// 		fmt.Println("Error:", err)
// 		return
// 	}

// 	encodedPerson := buf.Bytes()

// 	// fmt.Println("Gob encoded bytes:", encodedPerson)

// 	var decodedPerson Person
// 	dec := gob.NewDecoder(bytes.NewReader(encodedPerson))
// 	if err := dec.Decode(&decodedPerson); err != nil {
// 		fmt.Println("Error:", err)
// 		return
// 	}

// 	// fmt.Println("Gob decoded bytes:", decodedPerson)

// 	jsonData, err := json.Marshal(decodedPerson)
// 	if err != nil {
// 		fmt.Println("Error converting to JSON: ", err)
// 		return
// 	}

// 	// Convert the JSON byte slice to a string for display
// 	jsonString := string(jsonData)
// 	// fmt.Println("Decoded struct in JSON data format:", jsonData)
// 	fmt.Println("Decoded struct in JSON string format:", jsonString)
// }
