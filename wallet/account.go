package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
)

const (
	checkSumLength = 4
	version        = byte(0x00)
)

type Account struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

func (w Account) Address() []byte {

	pubHash := PublicKeyHash(w.PublicKey)

	versionedHash := append([]byte{version}, pubHash...)
	checkSum := CheckSum(versionedHash)

	fullHash := append(versionedHash, checkSum...)
	address := Base58Encode(fullHash)

	// fmt.Printf("\n*** >>> [pub key]: %x", w.PublicKey)
	// fmt.Printf("\n*** >>> [pub hash]: %x", pubHash)
	// fmt.Printf("\n*** >>> [address]: %s\n", address)

	fmt.Println()

	return address
}

func NewKeyPair() (ecdsa.PrivateKey, []byte) {

	curve := elliptic.P256()

	private, err := ecdsa.GenerateKey(curve, rand.Reader)

	Handle(err, "NewKeyPair")

	public := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, public
}

func MakeAccount() *Account {
	private, public := NewKeyPair()
	wallet := Account{private, public}
	return &wallet
}

func PublicKeyHash(pubKey []byte) []byte {

	pubHash := sha256.Sum256(pubKey)

	hasher := sha256.New()

	_, err := hasher.Write(pubHash[:])

	Handle(err, "PublicKeyHash")

	return hasher.Sum(nil)
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
