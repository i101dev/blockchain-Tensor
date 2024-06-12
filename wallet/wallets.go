package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/i101dev/blockchain-Tensor/util"
)

const walletFile = "../tmp/wallets.data"

type Wallet struct {
	Accounts map[string]*Account
}

func CreateWallets() (*Wallet, error) {

	wallet := Wallet{}
	wallet.Accounts = make(map[string]*Account)

	err := wallet.LoadFile()

	return &wallet, err
}

func (w *Wallet) Print() {
	fmt.Printf("\n%s", strings.Repeat("-", 48))
	fmt.Println("\nWallet Accounts:")
	fmt.Println()
	counter := 1
	for addr := range w.Accounts {
		fmt.Printf(" - Address %d: %s\n", counter, addr)
		counter++
	}
}

func (w Wallet) GetAccount(address string) Account {
	return *w.Accounts[address]
}

func (w *Wallet) AddAccount() string {

	account := MakeAccount()

	// addr := fmt.Sprintf("%s", account.Address())
	addr := string(account.Address())

	w.Accounts[addr] = account

	return addr
}

func (w *Wallet) GetAllAddresses() []string {

	var addresses []string

	for addr := range w.Accounts {
		addresses = append(addresses, addr)
	}

	return addresses
}

func (w *Wallet) LoadFile() error {

	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	var wallet Wallet

	fileContent, err := os.ReadFile(walletFile)
	util.Handle(err, "LoadFile 1")

	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallet)
	util.Handle(err, "LoadFile 2")

	w.Accounts = wallet.Accounts

	return nil
}

func (w *Wallet) SaveFile() {
	var content bytes.Buffer
	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(w)
	util.Handle(err, "SaveFile 1")

	err = os.WriteFile(walletFile, content.Bytes(), 0644)
	util.Handle(err, "SaveFile 2")
}

type PrivKey struct {
	D          *big.Int
	PublicKeyX *big.Int
	PublicKeyY *big.Int
}

func (w *Account) GobEncode() ([]byte, error) {
	privKey := &PrivKey{
		D:          w.PrivateKey.D,
		PublicKeyX: w.PrivateKey.PublicKey.X,
		PublicKeyY: w.PrivateKey.PublicKey.Y,
	}

	var buf bytes.Buffer

	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(privKey)
	if err != nil {
		return nil, err
	}

	_, err = buf.Write(w.PublicKey)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (w *Account) GobDecode(data []byte) error {
	buf := bytes.NewBuffer(data)
	var privKey PrivKey

	decoder := gob.NewDecoder(buf)
	err := decoder.Decode(&privKey)
	if err != nil {
		return err
	}

	w.PrivateKey = ecdsa.PrivateKey{
		D: privKey.D,
		PublicKey: ecdsa.PublicKey{
			X:     privKey.PublicKeyX,
			Y:     privKey.PublicKeyY,
			Curve: elliptic.P256(),
		},
	}
	w.PublicKey = make([]byte, buf.Len())
	_, err = buf.Read(w.PublicKey)
	if err != nil {
		return err
	}

	return nil
}
