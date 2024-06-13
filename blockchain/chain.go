package blockchain

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/i101dev/blockchain-Tensor/util"
	"github.com/i101dev/blockchain-Tensor/wallet"
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
		err := dbTXN.Set(newBlock.Hash, newBlock.Serialize())
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

// func (chain *Blockchain) AddBlock(txs []*Transaction) (*Block, error) {

// 	lastHash, err := chain.GetLastHash(chain.Database)
// 	if err != nil {
// 		return nil, err
// 	}

// 	newBlock, err := CreateBlock(txs, lastHash)
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = chain.PostBlockToDB(lastHash, newBlock, chain.Database)

//		return newBlock, err
//	}

func (chain *Blockchain) MineBlock(transactions []*Transaction) *Block {
	var lastHash []byte
	var lastHeight int

	for _, tx := range transactions {
		if !chain.VerifyTransaction(tx) {
			log.Panic("Invalid Transaction")
		}
	}

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		util.Handle(err, "MineBlock 1")
		lastHash, _ = item.ValueCopy(nil)

		item, err = txn.Get(lastHash)
		util.Handle(err, "MineBlock 2")
		lastBlockData, _ := item.ValueCopy(nil)

		lastBlock, _ := DeserializeBlock(lastBlockData)

		lastHeight = lastBlock.Height

		return err
	})

	util.Handle(err, "MineBlock 3")

	newBlock, _ := CreateBlock(transactions, lastHash, lastHeight+1)

	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		util.Handle(err, "MineBlock 4")
		err = txn.Set([]byte("lh"), newBlock.Hash)

		chain.LastHash = newBlock.Hash

		return err
	})
	util.Handle(err, "MineBlock 5")

	return newBlock
}

func (chain *Blockchain) AddBlock(block *Block) (*Block, error) {

	var b Block

	err := chain.Database.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get(block.Hash); err == nil {
			return nil
		}

		blockData := block.Serialize()
		err := txn.Set(block.Hash, blockData)
		util.Handle(err, "AddBlock 1")

		item, err := txn.Get([]byte("lh"))
		util.Handle(err, "AddBlock 2")
		lastHash, _ := item.ValueCopy(nil)

		item, err = txn.Get(lastHash)
		util.Handle(err, "AddBlock 3")
		lastBlockData, _ := item.ValueCopy(nil)

		lastBlock, _ := DeserializeBlock(lastBlockData)

		if block.Height > lastBlock.Height {
			err = txn.Set([]byte("lh"), block.Hash)
			util.Handle(err, "AddBlock 4")
			chain.LastHash = block.Hash
		}

		b = *lastBlock

		return nil
	})

	util.Handle(err, "AddBlock 5")

	return &b, err
}

func (chain *Blockchain) GetBlock(blockHash []byte) (*Block, error) {

	var block *Block

	err := chain.Database.View(func(txn *badger.Txn) error {
		if item, err := txn.Get(blockHash); err != nil {
			return errors.New("Block is not found")
		} else {
			blockData, _ := item.ValueCopy(nil)

			block, _ = DeserializeBlock(blockData)
		}
		return nil
	})
	if err != nil {
		return block, err
	}

	return block, nil
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
		block, err = DeserializeBlock(encodedBlock)

		return err
	})

	return block, err
}

func (chain *Blockchain) GetBlockHashes() [][]byte {

	var blocks [][]byte

	iter := chain.NewIterator()

	for {

		block, err := iter.IterateNext()
		if err != nil {
			break
		}

		blocks = append(blocks, block.Hash)

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return blocks
}

func (chain *Blockchain) GetBestHeight() int {

	var lastBlock *Block

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(LAST_HASH_KEY))
		util.Handle(err, "GetBestHeight 1")
		lastHash, _ := item.ValueCopy(nil)

		item, err = txn.Get(lastHash)
		util.Handle(err, "GetBestHeight 2")
		lastBlockData, _ := item.ValueCopy(nil)

		lastBlock, _ = DeserializeBlock(lastBlockData)

		return nil
	})

	util.Handle(err, "GetBestHeight 3")

	return lastBlock.Height
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

				if wallet.PubKeyHashToAddr(out.PubKeyHash) == address {
					utxoSet = append(utxoSet, &out)
				}
			}
		}
	}

	return utxoSet, nil
}

func (chain *Blockchain) FindTransaction(ID []byte) (Transaction, error) {

	iter := chain.NewIterator()

	for {
		block, err := iter.IterateNext()
		if err != nil {
			break
		}

		for _, tx := range block.Transactions {
			if bytes.Equal(tx.ID, ID) {
				return *tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction does not exist")
}

func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {

	prevTXs := make(map[string]Transaction)

	for _, txInput := range tx.Inputs {

		prevTX, err := bc.FindTransaction(txInput.ID)
		util.Handle(err, "SignTransaction")

		prevTXID := hex.EncodeToString(prevTX.ID)
		prevTXs[prevTXID] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {

	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {

		prevTX, err := bc.FindTransaction(in.ID)
		util.Handle(err, "VerifyTransaction")

		prevTXID := hex.EncodeToString(prevTX.ID)
		prevTXs[prevTXID] = prevTX
	}

	return tx.Verify(prevTXs)
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
			err = dbTXN.Set(genesis.Hash, genesis.Serialize())
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

	UTXOSet := UTXOSet{newChain}
	UTXOSet.Reindex()

	return newChain, err
}

func (chain *Blockchain) FindUTXO() map[string]TxOutputs {

	UTXO := make(map[string]TxOutputs)
	spentTXOs := make(map[string][]int)

	iter := chain.NewIterator()

	for {

		block, err := iter.IterateNext()
		if err != nil {
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

				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					inTxID := hex.EncodeToString(in.ID)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out)
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return UTXO
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

		block, err = DeserializeBlock(encodedBlock)
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
