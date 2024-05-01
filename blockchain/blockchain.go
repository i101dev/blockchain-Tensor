package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/dgraph-io/badger"
)

const (
	dbPath      = "./tmp/blocks"
	dbFile      = "./tmp/blocks/MANIFEST"
	genesisData = "First Transaction from Genesis"
)

type Blockchain struct {
	LastHash []byte
	Database *badger.DB
}

type BlockchainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

func InitBlockChain(address string) *Blockchain {

	if DBexists() {
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}

	var lastHash []byte

	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {

		genTxn := CoinbaseTx(address, genesisData)
		genesis := Genesis(genTxn)

		fmt.Println("Genesis proved")

		err = txn.Set(genesis.Hash, genesis.Serialize())

		Handle(err)

		err = txn.Set([]byte("lh"), genesis.Hash)

		lastHash = genesis.Hash

		return err
	})

	Handle(err)

	return &Blockchain{lastHash, db}
}

func ContinueBlockChain(address string) *Blockchain {

	if !DBexists() {
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}

	var lastHash []byte

	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {

		item, err := txn.Get([]byte("lh"))

		Handle(err)

		lastHash, err = item.ValueCopy(nil)

		return err
	})

	Handle(err)

	return &Blockchain{lastHash, db}
}

func DBexists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

func (chain *Blockchain) AddBlock(transactions []*Transaction) *Block {

	var lasthash []byte

	err := chain.Database.View(func(txn *badger.Txn) error {

		item, err := txn.Get([]byte("lh"))

		Handle(err)

		lasthash, err = item.ValueCopy(nil)

		return err

	})

	Handle(err)

	newBlock := CreateBlock(transactions, lasthash)

	err = chain.Database.Update(func(txn *badger.Txn) error {

		err := txn.Set(newBlock.Hash, newBlock.Serialize())

		Handle(err)

		err = txn.Set([]byte("lh"), newBlock.Hash)

		chain.LastHash = newBlock.Hash

		return err
	})

	Handle(err)

	return newBlock
}

func (chain *Blockchain) Iterator() *BlockchainIterator {

	iter := &BlockchainIterator{chain.LastHash, chain.Database}

	return iter
}

func (iter *BlockchainIterator) NextBlock() *Block {

	var block *Block

	err := iter.Database.View(func(txn *badger.Txn) error {

		item, err := txn.Get(iter.CurrentHash)

		Handle(err)

		encodedBlock, err := item.ValueCopy(nil)

		block = Deserialize(encodedBlock)

		return err
	})

	Handle(err)

	iter.CurrentHash = block.PrevHash

	return block
}

// func (chain *Blockchain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {

// 	var unspentTxs []Transaction

// 	spentTXOs := make(map[string][]int)

// 	iter := chain.Iterator()

// 	for {
// 		block := iter.NextBlock()

// 		for _, tx := range block.Transactions {
// 			txID := hex.EncodeToString(tx.HashID)

// 		Outputs:
// 			for outIdx, out := range tx.Outputs {
// 				if spentTXOs[txID] != nil {
// 					for _, spentOut := range spentTXOs[txID] {
// 						if spentOut == outIdx {
// 							continue Outputs
// 						}
// 					}
// 				}

// 				if out.IsLockedWithKey(pubKeyHash) {
// 					unspentTxs = append(unspentTxs, *tx)
// 				}
// 			}

// 			if !tx.IsCoinbase() {
// 				for _, in := range tx.Inputs {
// 					if in.UsesKey(pubKeyHash) {
// 						inTxID := hex.EncodeToString(in.HashID)
// 						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out)
// 					}
// 				}
// 			}
// 		}

// 		if len(block.PrevHash) == 0 {
// 			break
// 		}
// 	}

// 	return unspentTxs
// }

func (chain *Blockchain) FindUTXO() map[string]TxOutputs {

	UTXO := make(map[string]TxOutputs)
	spentTXOs := make(map[string][]int)

	iter := chain.Iterator()

	for {
		block := iter.NextBlock()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.HashID)

		Outputs:
			for outIdx, out := range tx.Outputs {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					inTxId := hex.EncodeToString(in.HashID)
					spentTXOs[inTxId] = append(spentTXOs[inTxId], in.Out)
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return UTXO
}

// func (chain *Blockchain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {

// 	unspentOuts := make(map[string][]int)
// 	unspentTxs := chain.FindUnspentTransactions(pubKeyHash)
// 	accumulated := 0

// Work:
// 	for _, tx := range unspentTxs {
// 		txID := hex.EncodeToString(tx.HashID)

// 		for outIdx, out := range tx.Outputs {
// 			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
// 				accumulated += out.Value
// 				unspentOuts[txID] = append(unspentOuts[txID], outIdx)

// 				if accumulated >= amount {
// 					break Work
// 				}
// 			}
// 		}
// 	}

// 	return accumulated, unspentOuts
// }

func (chain *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	iter := chain.Iterator()

	for {
		block := iter.NextBlock()

		for _, tx := range block.Transactions {
			if bytes.Equal(tx.HashID, ID) {
				return *tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction does not exist")
}

func (chain *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {

	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTX, err := chain.FindTransaction(in.HashID)
		Handle(err)
		prevTXs[hex.EncodeToString(prevTX.HashID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

func (chain *Blockchain) VerifyTransaction(tx *Transaction) bool {

	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTX, err := chain.FindTransaction(in.HashID)
		Handle(err)
		prevTXs[hex.EncodeToString(prevTX.HashID)] = prevTX
	}

	return tx.Verify(prevTXs)
}
