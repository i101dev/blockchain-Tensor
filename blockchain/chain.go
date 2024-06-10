package blockchain

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/i101dev/blockchain-Tensor/types"
	"github.com/i101dev/blockchain-Tensor/util"
)

const (
	DB_PATH   = "../tmp/blocks_"
	LAST_HASH = "lastHash"
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

func OpenDB(chain *Blockchain) *badger.DB {

	opts := badger.DefaultOptions(chain.Path)
	opts.Logger = &NullLogger{}

	db, err := badger.Open(opts)
	util.Handle(err, "Open BadgerDB 1")

	chain.Database = db

	return db
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

func (chain *Blockchain) GetLastHash(db *badger.DB) []byte {

	var lastHash []byte

	db.View(func(txn *badger.Txn) error {

		item, err := txn.Get([]byte(LAST_HASH))
		util.Handle(err, "GetLastHash 1")

		lastHash, err = item.ValueCopy(nil)
		util.Handle(err, "GetLastHash 2")

		return nil
	})

	return lastHash
}

func (chain *Blockchain) PostBlockToDB(lastHash []byte, newBlock *Block, db *badger.DB) {

	db.Update(func(txn *badger.Txn) error {

		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		util.Handle(err, "AddBlock 3")

		err = txn.Set([]byte(LAST_HASH), newBlock.Hash)
		util.Handle(err, "AddBlock 4")

		chain.LastHash = newBlock.Hash

		return nil
	})
}

func (chain *Blockchain) AddBlock(data *types.AddBlockReq) {

	// fmt.Printf("Incoming data: %+v\n", *data.Data)

	db := OpenDB(chain)
	defer chain.CloseDB()

	lastHash := chain.GetLastHash(db)
	newBlock := CreateBlock(*data.Data, lastHash)

	chain.PostBlockToDB(lastHash, newBlock, db)
}

func (chain *Blockchain) GetBlockByHash(db *badger.DB, hash []byte) *Block {

	var block *Block

	db.View(func(txn *badger.Txn) error {

		item, err := txn.Get(hash)
		util.Handle(err, "GetBlockByHash - 1")

		encodedBlock, err := item.ValueCopy(nil)
		util.Handle(err, "GetBlockByHash - 2")

		block = Deserialize(encodedBlock)

		return nil
	})

	return block
}

func InitBlockchain(address string, nodeID uint16) *Blockchain {

	path := fmt.Sprintf("%s%d", DB_PATH, nodeID)

	// Ensure the directory exists
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		log.Panicf(fmt.Sprintf("Error Creating Dir: %s", err))
	}

	// -------------------------------------------------

	newChain := &Blockchain{
		Path: path,
	}

	db := OpenDB(newChain)
	defer newChain.CloseDB()

	var lastHash []byte
	db.Update(func(txn *badger.Txn) error {

		if _, err := txn.Get([]byte(LAST_HASH)); err != nil {

			genesis := Genesis()
			fmt.Println("Starting new chain - Genesis proved")

			// Using Genesis block has as the [Key]
			// Set the serialized Genesis block as the [Value]
			err = txn.Set(genesis.Hash, genesis.Serialize())
			util.Handle(err, "InitBlockchain 1")

			// Set the hash from the Genesis block
			// as the last and most recent hash for the entire chain
			err = txn.Set([]byte(LAST_HASH), genesis.Hash)
			util.Handle(err, "InitBlockchain 2")

			lastHash = genesis.Hash

			return nil

		} else {

			item, err := txn.Get([]byte(LAST_HASH))
			util.Handle(err, "InitBlockchain 3")

			lastHash, err = item.ValueCopy(nil)
			util.Handle(err, "InitBlockchain 4")

			return nil
		}
	})

	newChain.LastHash = lastHash

	return newChain
}

func DBexists(path string) bool {
	if _, err := os.Stat(path + "/MANIFEST"); os.IsNotExist(err) {
		return false
	}
	return true
}

// -----------------------------------------------------------------------
type BlockchainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
	Chain       *Blockchain
}

func (chain *Blockchain) NewIterator() *BlockchainIterator {
	return &BlockchainIterator{
		CurrentHash: chain.LastHash,
		Database:    chain.Database,
		Chain:       chain,
	}
}

func (iter *BlockchainIterator) IterateNext() *Block {

	block := iter.Chain.GetBlockByHash(iter.Database, iter.CurrentHash)

	iter.CurrentHash = block.PrevHash

	return block
}
