package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dgraph-io/badger"
)

const (
	dbPath      = "./tmp/blocks_%s"
	genesisData = "First Transaction from Genesis"
)

type Blockchain struct {
	LastHash []byte
	Database *badger.DB
}

type NullLogger struct{}

func (l *NullLogger) Errorf(string, ...interface{})   {}
func (l *NullLogger) Warningf(string, ...interface{}) {}
func (l *NullLogger) Infof(string, ...interface{})    {}
func (l *NullLogger) Debugf(string, ...interface{})   {}

// --------------------------------------------------------------------
// --------------------------------------------------------------------

func InitBlockChain(address string, nodeID string) *Blockchain {

	path := fmt.Sprintf(dbPath, nodeID)

	// fmt.Printf("\nPath %s\n", path)
	// fmt.Println()

	if DBexists(path) {
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}

	var lastHash []byte

	opts := badger.DefaultOptions(path)
	opts.Logger = &NullLogger{}

	db, err := openDB(path, opts)
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

func ContinueBlockChain(nodeID string) *Blockchain {

	path := fmt.Sprintf(dbPath, nodeID)

	fmt.Println()

	if !DBexists(path) {
		fmt.Println("Blockchain doesn't exist!")
		runtime.Goexit()
	}

	var lastHash []byte

	opts := badger.DefaultOptions(path)
	opts.Logger = &NullLogger{}
	db, err := openDB(path, opts)
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

func DBexists(path string) bool {
	if _, err := os.Stat(path + "/MANIFEST"); os.IsNotExist(err) {
		return false
	}
	return true
}

func (chain *Blockchain) AddBlock(block *Block) {

	err := chain.Database.Update(func(txn *badger.Txn) error {

		if _, err := txn.Get(block.Hash); err == nil {
			return nil
		}

		blockData := block.Serialize()
		err := txn.Set(block.Hash, blockData)
		Handle(err)

		item, err := txn.Get([]byte("lh"))
		Handle(err)

		lastHash, _ := item.ValueCopy(nil)
		item, err = txn.Get(lastHash)
		Handle(err)

		lastBlockData, _ := item.ValueCopy(nil)
		lastBlock := Deserialize(lastBlockData)

		if block.Height > lastBlock.Height {
			err = txn.Set([]byte("lh"), block.Hash)
			Handle(err)
			chain.LastHash = block.Hash
		}

		return nil
	})

	Handle(err)
}

func (chain *Blockchain) GetBlock(blockHash []byte) (Block, error) {

	var block Block

	err := chain.Database.View(func(txn *badger.Txn) error {

		if item, err := txn.Get(blockHash); err != nil {
			return errors.New("Block is not found")

		} else {

			blockData, _ := item.ValueCopy(nil)
			block = *Deserialize(blockData)
		}

		return nil
	})

	if err != nil {
		return block, err
	}

	return block, nil
}

func (chain *Blockchain) GetBlockHashes() [][]byte {

	var blocks [][]byte

	iter := chain.Iterator()

	for {

		block := iter.NextBlock()

		blocks = append(blocks, block.Hash)

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return blocks
}

func (chain *Blockchain) GetBestHeight() int {

	var lastBlock Block

	err := chain.Database.View(func(txn *badger.Txn) error {

		item, err := txn.Get([]byte("lh"))
		Handle(err)

		lastHash, _ := item.ValueCopy(nil)
		item, err = txn.Get(lastHash)
		Handle(err)

		lastBlockData, _ := item.ValueCopy(nil)
		lastBlock = *Deserialize(lastBlockData)

		return nil
	})

	Handle(err)

	return lastBlock.Height
}

func (chain *Blockchain) MineBlock(transactions []*Transaction) *Block {

	var lastHash []byte
	var lastHeight int

	err := chain.Database.View(func(txn *badger.Txn) error {

		item, err := txn.Get([]byte("lh"))
		Handle(err)

		lastHash, err = item.ValueCopy(nil)
		Handle(err)

		item, err = txn.Get(lastHash)
		Handle(err)

		lastBlockData, _ := item.ValueCopy(nil)
		lastBlock := Deserialize(lastBlockData)
		lastHeight = lastBlock.Height

		return err
	})

	Handle(err)

	newBlock := CreateBlock(transactions, lastHash, lastHeight+1)

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

func (chain *Blockchain) FindUTXO() map[string]TxOutputs {

	UTXO := make(map[string]TxOutputs)
	spentTXOs := make(map[string][]int)

	iter := chain.Iterator()

	for {
		block := iter.NextBlock()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

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

func (chain *Blockchain) FindTransaction(ID []byte) (Transaction, error) {

	iter := chain.Iterator()

	for {

		block := iter.NextBlock()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
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
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

func (chain *Blockchain) VerifyTransaction(tx *Transaction) bool {

	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTX, err := chain.FindTransaction(in.HashID)
		Handle(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}

func retry(dir string, originalOpts badger.Options) (*badger.DB, error) {

	lockpath := filepath.Join(dir, "LOCK")

	if err := os.Remove(lockpath); err != nil {
		return nil, fmt.Errorf(`removing "LOCK": %s`, err)
	}

	retryOpts := originalOpts
	retryOpts.Truncate = true
	retryOpts.Logger = &NullLogger{}

	db, err := badger.Open(retryOpts)

	return db, err
}

func openDB(dir string, opts badger.Options) (*badger.DB, error) {

	if db, err := badger.Open(opts); err != nil {

		if strings.Contains(err.Error(), "LOCK") {

			if db, err := retry(dir, opts); err == nil {
				log.Println("*** >>> database unlocked, value log truncated")
				return db, nil
			}

			log.Println("*** >>> could not unlock database")
		}

		return nil, err

	} else {

		return db, nil
	}
}
