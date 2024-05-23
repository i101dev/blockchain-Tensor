package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"log"
	"math/big"
	"os"
)

const walletFile = "./tmp/wallets_%s.data"

type Wallet struct {
	Accounts map[string]*Account
}

func LoadWallet(nodeID string) (*Wallet, error) {
	wallets := Wallet{}
	wallets.Accounts = make(map[string]*Account)
	err := wallets.LoadFile(nodeID)
	return &wallets, err
}

func (ws *Wallet) AddAccount() string {
	wallet := MakeAccount()
	address := string(wallet.Address())

	ws.Accounts[address] = wallet

	return address
}

func (ws *Wallet) GetAllAddresses() []string {
	var addresses []string
	for address := range ws.Accounts {
		addresses = append(addresses, address)
	}
	return addresses
}

func (ws Wallet) GetAccount(address string) Account {
	return *ws.Accounts[address]
}

func (ws *Wallet) LoadFile(nodeID string) error {

	walletFile := fmt.Sprintf(walletFile, nodeID)
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	fileContent, err := os.ReadFile(walletFile)
	if err != nil {
		return err
	}

	var w Wallet
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&w)
	if err != nil {
		return err
	}

	ws.Accounts = w.Accounts

	return nil
}

func (ws *Wallet) SaveFile(nodeID string) {

	var content bytes.Buffer
	walletFile := fmt.Sprintf(walletFile, nodeID)

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		log.Panic(err)
	}

	err = os.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}

// ----------------------------------------------------------
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
