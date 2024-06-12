package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"

	"github.com/i101dev/blockchain-Tensor/util"
	"golang.org/x/crypto/ripemd160"
)

// -----------------------------------------------------------------------
const (
	checksumLength = 4
	version        = byte(0x00)
)

// -----------------------------------------------------------------------

type Account struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

func (w Account) Address() []byte {

	publicHash := PublicKeyHash(w.PublicKey)

	versionedHash := append([]byte{version}, publicHash...)
	checkSum := CheckSum(versionedHash)

	fullHash := append(versionedHash, checkSum...)
	address := util.Base58Encode(fullHash)

	return address
}

func PubKeyHashToAddr(pubKeyHash []byte) string {

	versionedHash := append([]byte{version}, pubKeyHash...)
	checkSum := CheckSum(versionedHash)

	fullHash := append(versionedHash, checkSum...)
	address := util.Base58Encode(fullHash)

	return string(address)
}

func NewKeyPair() (ecdsa.PrivateKey, []byte) {

	curve := elliptic.P256()

	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	public := append(private.X.Bytes(), private.Y.Bytes()...)

	return *private, public
}

func MakeAccount() *Account {
	private, public := NewKeyPair()
	return &Account{
		PrivateKey: private,
		PublicKey:  public,
	}
}

func PublicKeyHash(pubKey []byte) []byte {

	pubHash := sha256.Sum256(pubKey)
	hasher := ripemd160.New()

	if _, err := hasher.Write(pubHash[:]); err != nil {
		log.Fatal(err)
	}

	pubRipMD := hasher.Sum(nil)

	return pubRipMD
}

func CheckSum(payload []byte) []byte {

	hashOne := sha256.Sum256(payload)
	hashTwo := sha256.Sum256(hashOne[:])

	return hashTwo[:checksumLength]
}

func ValidateAddress(address string) bool {

	pubKeyHash := util.Base58Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-checksumLength:]

	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-checksumLength]

	targetChecksum := CheckSum(append([]byte{version}, pubKeyHash...))

	return bytes.Equal(actualChecksum, targetChecksum)
}

// -----------------------------------------------------------------------
