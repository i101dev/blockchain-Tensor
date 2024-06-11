package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
)

const Difficulty = 12

type ProofOfWork struct {
	Block  *Block
	Target *big.Int
}

func ToHex(num int64) ([]byte, error) {

	buff := new(bytes.Buffer)

	if err := binary.Write(buff, binary.BigEndian, num); err != nil {
		return []byte{}, err
	}

	return buff.Bytes(), nil
}

func NewProof(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-Difficulty))
	pow := &ProofOfWork{b, target}
	return pow
}

func (pow *ProofOfWork) InitData(nonce int) ([]byte, error) {

	non, err := ToHex(int64(nonce))
	if err != nil {
		return nil, fmt.Errorf("failed to write [nonce] ToHex")
	}

	diff, err := ToHex(int64(Difficulty))
	if err != nil {
		return nil, fmt.Errorf("failed to write [difficulty] ToHex")
	}

	data := bytes.Join(
		[][]byte{
			pow.Block.PrevHash,
			pow.Block.HashTransactions(),
			non,
			diff,
			// ToHex(int64(nonce)),
			// ToHex(int64(Difficulty)),
		},
		[]byte{},
	)

	return data, nil
}

func (pow *ProofOfWork) Validate() (bool, error) {

	var intHash big.Int

	data, err := pow.InitData(pow.Block.Nonce)
	if err != nil {
		return false, err
	}

	hash := sha256.Sum256(data)

	intHash.SetBytes(hash[:])

	return intHash.Cmp(pow.Target) == -1, nil
}

func (pow *ProofOfWork) Run() (int, []byte, error) {

	var intHash big.Int
	var hash [32]byte

	nonce := 0

	for nonce < math.MaxInt64 {

		data, err := pow.InitData(nonce)
		if err != nil {
			return -1, nil, err
		}

		hash = sha256.Sum256(data)

		intHash.SetBytes(hash[:])

		if intHash.Cmp(pow.Target) == -1 {
			break
		} else {
			nonce++
		}
	}

	return nonce, hash[:], nil
}
