package wallet2

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/btcsuite/btcutil/base58"
)

const (
	checkSumLength = 4
	version        = 0x00
	walletFile     = "./tmp/wallets_%s.data"
)

type Wallet struct {
	Accounts map[string]*Account
}
type Account struct {
	privateKey  ecdsa.PrivateKey
	publicKey   ecdsa.PublicKey
	pubKeyBytes []byte
	address     string
}

// ----------------------------------------------------------------------------
// [Account] Methods
func NewAccount() *Account {

	// 1. Creating ECDSA private key (32 bytes) & public key (64 byte)
	w := new(Account)
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	w.privateKey = *privateKey
	w.publicKey = w.privateKey.PublicKey
	w.pubKeyBytes = append(w.privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...)

	// 2. Perform SHA-256 hashing on the public key (32 bytes)
	h2 := sha256.New()
	h2.Write(w.publicKey.X.Bytes())
	h2.Write(w.publicKey.Y.Bytes())
	digest2 := h2.Sum(nil)

	// 3. Perform RIPEMD-160 hashing on the result of SHA-256 (20 bytes)
	h3 := sha256.New()
	h3.Write(digest2)
	digest3 := h3.Sum(nil)

	// 4. Add version byte in from of RIPEMD-160 hash (0x00 for main network)
	vd4 := make([]byte, 21)
	vd4[0] = version
	copy(vd4[1:], digest3[:])

	// 5. Perform SHA-256 hash on the extended RIPEMD-160 result
	h5 := sha256.New()
	h5.Write(vd4)
	digest5 := h5.Sum(nil)

	// 6. Perform SHA-256 hash on the result of the previous SHA-256 hash
	h6 := sha256.New()
	h6.Write(digest5)
	digest6 := h6.Sum(nil)

	// 7. Take the first 4 bytes of the second SHA-256 hash checksum
	checkSum := digest6[:checkSumLength]

	// 8. Add the 4 checksum bytes from 7 at the end of the extended RIPEMD-160 hash from step #4 (25 bytes)
	dc8 := make([]byte, 25)
	copy(dc8[:21], vd4[:])
	copy(dc8[21:], checkSum[:])

	// 9. Convert the result from a byte in to base58
	w.address = base58.Encode(dc8)

	return w
}
func PublicKeyHash(pubKey []byte) []byte {

	pubHash := sha256.Sum256(pubKey)

	hasher := sha256.New()

	_, err := hasher.Write(pubHash[:])

	Handle(err, "PublicKeyHash")

	return hasher.Sum(nil)
}
func (acct *Account) PrivateKey() ecdsa.PrivateKey {
	return acct.privateKey
}
func (acct *Account) PrivateKeyStr() string {
	return fmt.Sprintf("%x", acct.privateKey.D.Bytes())
}
func (acct *Account) PublicKey() ecdsa.PublicKey {
	return acct.publicKey
}
func (acct *Account) PublicKeyStr() string {
	return fmt.Sprintf("%x%x", acct.publicKey.X.Bytes(), acct.publicKey.Y.Bytes())
}
func (acct *Account) PublicKeyBytes() []byte {
	return acct.pubKeyBytes
}
func (acct *Account) Address() string {
	return acct.address
}
func (acct *Account) GobEncode() ([]byte, error) {

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	err := encoder.Encode(acct.privateKey.D)
	if err != nil {
		return nil, err
	}

	err = encoder.Encode(acct.privateKey.PublicKey.X)
	if err != nil {
		return nil, err
	}

	err = encoder.Encode(acct.privateKey.PublicKey.Y)
	if err != nil {
		return nil, err
	}

	err = encoder.Encode(acct.address)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
func (acct *Account) GobDecode(data []byte) error {

	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)

	var d *big.Int
	var x, y *big.Int

	err := decoder.Decode(&d)
	if err != nil {
		return err
	}

	err = decoder.Decode(&x)
	if err != nil {
		return err
	}

	err = decoder.Decode(&y)
	if err != nil {
		return err
	}

	err = decoder.Decode(&acct.address)
	if err != nil {
		return err
	}

	acct.privateKey = ecdsa.PrivateKey{
		D: d,
		PublicKey: ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     x,
			Y:     y,
		},
	}

	acct.publicKey = acct.privateKey.PublicKey

	return nil
}

// ----------------------------------------------------------------------------
// [Wallet] Methods
func MakeWallet() *Wallet {
	return &Wallet{
		Accounts: make(map[string]*Account),
	}
}
func CreateWallet(nodeID string) (*Wallet, error) {
	wallet := MakeWallet()
	wallet.AddAccount()
	wallet.SaveFile(nodeID)
	return wallet, nil
}
func LoadWallet(nodeID string) (*Wallet, error) {
	wallet := MakeWallet()
	err := wallet.LoadFile(nodeID)
	return wallet, err
}
func (w *Wallet) AddAccount() string {
	account := NewAccount()
	address := string(account.Address())

	w.Accounts[address] = account

	return address
}
func (w *Wallet) GetAccount(address string) *Account {
	return w.Accounts[address]
}
func (w *Wallet) GetAllAddresses() []string {
	var addresses []string
	for address := range w.Accounts {
		addresses = append(addresses, address)
	}
	return addresses
}
func (w *Wallet) LoadFile(nodeID string) error {

	walletFile := fmt.Sprintf(walletFile, nodeID)
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	fileContent, err := os.ReadFile(walletFile)
	if err != nil {
		return err
	}

	var wallet Wallet
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallet)
	if err != nil {
		return err
	}

	w.Accounts = wallet.Accounts

	return nil
}
func (w *Wallet) SaveFile(nodeID string) {

	var content bytes.Buffer
	walletFile := fmt.Sprintf(walletFile, nodeID)

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(w)
	if err != nil {
		log.Panic(err)
	}

	err = os.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}

// ----------------------------------------------------------------------------
// Utility functions
func Handle(err error, functionName string) {
	if err != nil {
		fmt.Printf(`\n*** >>> ERROR @ [%s] - %+v`, functionName, err)
	}
}
func Base58Encode(input []byte) []byte {
	encode := base58.Encode(input)
	return []byte(encode)
}
func Base58Decode(input []byte) []byte {
	decode := base58.Decode(string(input[:]))
	return decode
}
func CheckSum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])
	return secondHash[:checkSumLength]
}
func ValidAddress(address string) bool {

	pubKeyHash := Base58Decode([]byte(address))
	actualCheckSum := pubKeyHash[len(pubKeyHash)-checkSumLength:]

	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-checkSumLength]
	targetCheckSum := CheckSum(append([]byte{version}, pubKeyHash...))

	return bytes.Equal(actualCheckSum, targetCheckSum)
}

// ----------------------------------------------------------------------------

func Test() {

	testFile := "JIMBO"

	wallet, err := CreateWallet(testFile)
	if err != nil {
		log.Fatal("\n*** >>> [CreateWallet] - FATAL -", err)
	}

	allAddrs := wallet.GetAllAddresses()

	fmt.Println("\nwallet -", allAddrs)

	acct := wallet.GetAccount(allAddrs[0])

	fmt.Printf("\n*** >>> [PublicKeyBytes] - %+v", acct.PublicKeyBytes())
	// fmt.Printf("\n*** >>> [PrivateKey] - %+v", acct.PrivateKey())
	// fmt.Printf("\n*** >>> [PrivateKeyStr] - %+v", acct.PrivateKeyStr())
	// fmt.Printf("\n*** >>> [PublicKey] - %+v", acct.PublicKey())
	// fmt.Printf("\n*** >>> [PublicKeyStr] - %+v", acct.PublicKeyStr())
	// fmt.Printf("\n*** >>> [Address] - %+v\n", acct.Address())

}
