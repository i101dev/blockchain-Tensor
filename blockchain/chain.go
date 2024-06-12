package blockchain

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/i101dev/blockchain-Tensor/util"
)

const (
	DB_PATH       = "../tmp/blocks_%d"
	LAST_HASH_KEY = "lastHash"
	GENESIS_DATA  = "GENESIS"
)

type NullLogger struct{}

func (l *NullLogger) Errorf(string, ...interface{})   {}
func (l *NullLogger) Warningf(string, ...interface{}) {}
func (l *NullLogger) Infof(string, ...interface{})    {}
func (l *NullLogger) Debugf(string, ...interface{})   {}

type Blockchain struct {
	Path     string
	LastHash []byte
	Database *badger.DB
}

func (chain *Blockchain) CloseDB() {

	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for {
		err := chain.Database.RunValueLogGC(0.5)
		if err != nil {
			break
		}
	}

	err := chain.Database.Close()
	util.Handle(err, "Close 1")
}

func (chain *Blockchain) GetLastHash(db *badger.DB) ([]byte, error) {

	var lastHash []byte

	err := db.View(func(txn *badger.Txn) error {

		// ----------------------------------------------------------
		item, err := txn.Get([]byte(LAST_HASH_KEY))
		if err != nil {
			return fmt.Errorf("failed to get last hash from bytes")
		}

		// ----------------------------------------------------------
		lastHash, err = item.ValueCopy(nil)
		if err != nil {
			return fmt.Errorf("failed item ValueCopy")
		}

		return nil
	})

	return lastHash, err
}

func (chain *Blockchain) PostBlockToDB(lastHash []byte, newBlock *Block, db *badger.DB) error {

	return db.Update(func(dbTXN *badger.Txn) error {

		// ----------------------------------------------------------
		serializedBlock, err := newBlock.Serialize()
		if err != nil {
			return err
		}

		// ----------------------------------------------------------
		err = dbTXN.Set(newBlock.Hash, serializedBlock)
		if err != nil {
			return fmt.Errorf("failed to set serialized block in database")
		}

		// ----------------------------------------------------------
		err = dbTXN.Set([]byte(LAST_HASH_KEY), newBlock.Hash)
		if err != nil {
			return fmt.Errorf("failed to set LAST_HASH in database")
		}

		chain.LastHash = newBlock.Hash

		return nil
	})
}

func (chain *Blockchain) AddBlock(txs []*Transaction) error {

	lastHash, err := chain.GetLastHash(chain.Database)
	if err != nil {
		return err
	}

	newBlock, err := CreateBlock(txs, lastHash)
	if err != nil {
		return err
	}

	return chain.PostBlockToDB(lastHash, newBlock, chain.Database)
}

func (chain *Blockchain) GetAllBlocks() []*Block {

	iter := chain.NewIterator()

	var allBlocks []*Block

	for {

		block, err := iter.IterateNext()
		if err != nil {
			break
		}

		allBlocks = append(allBlocks, block)

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return allBlocks
}

func (chain *Blockchain) GetBlockByHash(db *badger.DB, hash []byte) (*Block, error) {

	var block *Block

	err := db.View(func(txn *badger.Txn) error {

		// ----------------------------------------------------------
		item, err := txn.Get(hash)
		if err != nil {
			return fmt.Errorf("HASH NOT FOUND")
		}

		// ----------------------------------------------------------
		encodedBlock, err := item.ValueCopy(nil)
		if err != nil {
			return fmt.Errorf("FAILED TO ENCODE")
		}

		// ----------------------------------------------------------
		block, err = Deserialize(encodedBlock)

		return err
	})

	return block, err
}

func (chain *Blockchain) GetUnspentOutputs(db *badger.DB, address string) ([]*TxOutput, error) {

	var utxoSet []*TxOutput

	iter := chain.NewIterator()

	for {

		block, err := iter.IterateNext()
		if err != nil {
			break
		}

		for _, tx := range block.Transactions {

			for _, out := range tx.Outputs {

				if out.CanBeUnlocked(address) {
					utxoSet = append(utxoSet, &out)
				}
			}
		}
	}

	return utxoSet, nil
}

// -----------------------------------------------------------------------
func OpenDB(chain *Blockchain) *badger.DB {

	opts := badger.DefaultOptions(chain.Path)
	opts.Logger = &NullLogger{}

	db, err := badger.Open(opts)
	util.Handle(err, "Open BadgerDB 1")

	chain.Database = db

	return db
}

func NewChain(nodeID uint16) *Blockchain {

	path := fmt.Sprintf(DB_PATH, nodeID)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		log.Fatalf(fmt.Sprintf("Error Creating Dir: %s", err))
	}

	return &Blockchain{
		Path: path,
	}
}

func LoadBlockchain(address string, nodeID uint16) (*Blockchain, error) {

	path := fmt.Sprintf(DB_PATH, nodeID)

	// Ensure the directory exists ---------------------------
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		log.Panicf(fmt.Sprintf("Error Creating Dir: %s", err))
	}

	newChain := &Blockchain{
		Path: path,
	}

	// -------------------------------------------------------
	db := OpenDB(newChain)
	defer newChain.CloseDB()

	var lastHash []byte
	err := db.Update(func(dbTXN *badger.Txn) error {

		if _, err := dbTXN.Get([]byte(LAST_HASH_KEY)); err == badger.ErrKeyNotFound {

			// ----------------------------------------------------------
			cbtx := CoinbaseTX(address, GENESIS_DATA)
			genesis, err := Genesis(cbtx)
			if err != nil {
				return err
			}

			// ----------------------------------------------------------
			serializedBlock, err := genesis.Serialize()
			if err != nil {
				return err
			}

			// ----------------------------------------------------------
			err = dbTXN.Set(genesis.Hash, serializedBlock)
			if err != nil {
				return fmt.Errorf("failed to set serialized block in database")
			}

			// ----------------------------------------------------------
			err = dbTXN.Set([]byte(LAST_HASH_KEY), genesis.Hash)
			if err != nil {
				return fmt.Errorf("failed to set LAST_HASH in database")
			}

			lastHash = genesis.Hash

			return nil
		}

		last, err := newChain.GetLastHash(db)

		lastHash = last

		return err
	})

	newChain.LastHash = lastHash

	return newChain, err
}

// -----------------------------------------------------------------------
// UTXO - Tensor methods

func (chain *Blockchain) FindUnspentTransactions(address string) []Transaction {
	var unspentTxs []Transaction

	spentTXOs := make(map[string][]int)

	iter := chain.NewIterator()

	for {
		block, _ := iter.IterateNext()

		if block == nil {
			break
		}

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
				if out.CanBeUnlocked(address) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}
			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					if in.CanUnlock(address) {
						inTxID := hex.EncodeToString(in.ID)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out)
					}
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return unspentTxs
}

func (chain *Blockchain) FindUTXO(address string) []TxOutput {

	unspentTransactions := chain.FindUnspentTransactions(address)

	var UTXOs []TxOutput
	for _, tx := range unspentTransactions {
		for _, out := range tx.Outputs {
			if out.CanBeUnlocked(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

func (chain *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {

	accumulated := 0
	unspentOuts := make(map[string][]int)
	unspentTxs := chain.FindUnspentTransactions(address)

	for _, tx := range unspentTxs {

		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Outputs {

			if out.CanBeUnlocked(address) && accumulated < amount {

				accumulated += out.Value
				unspentOuts[txID] = append(unspentOuts[txID], outIdx)

				if accumulated >= amount {
					return accumulated, unspentOuts
				}
			}
		}
	}

	return accumulated, unspentOuts
}

// -----------------------------------------------------------------------
type BlockchainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
	Chain       *Blockchain
}

func (iter *BlockchainIterator) Print() {

	block, err := iter.IterateNext()
	if err != nil {
		return
	}

	fmt.Printf("Current Hash: %x\n", block.Hash)
	fmt.Printf("Previous Hash: %x\n", block.PrevHash)
	fmt.Println("Transactions:")
	for _, tx := range block.Transactions {
		tx.Print()
	}
}

func (chain *Blockchain) NewIterator() *BlockchainIterator {
	return &BlockchainIterator{
		CurrentHash: chain.LastHash,
		Database:    chain.Database,
		Chain:       chain,
	}
}

func (iter *BlockchainIterator) IterateNext() (*Block, error) {
	var block *Block

	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		if err != nil {
			return fmt.Errorf(err.Error())
		}

		encodedBlock, err := item.ValueCopy(nil)
		if err != nil {
			return fmt.Errorf(err.Error())
		}

		block, err = Deserialize(encodedBlock)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if block == nil {
		return nil, fmt.Errorf("retrieved block is nil")
	}

	iter.CurrentHash = block.PrevHash

	return block, nil
}
