package blockchain

import (
	"fmt"

	"github.com/dgraph-io/badger"
)

const dbPath = "./tmp/blocks"

type Blockchain struct {
	LastHash []byte
	Database *badger.DB
}

type BlockchainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

func InitBlockChain() *Blockchain {

	var lastHash []byte

	opts := badger.DefaultOptions(dbPath)

	db, err := badger.Open(opts)

	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {

		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {

			fmt.Println("No existing blockchain found")

			genesis := Genesis()

			fmt.Println("Genesis proved")

			err = txn.Set(genesis.Hash, genesis.Serialize())

			Handle(err)

			err = txn.Set([]byte("lh"), genesis.Hash)

			lastHash = genesis.Hash

			return err

		} else {

			fmt.Println("Continuing existing chain")

			item, err := txn.Get([]byte("lh"))

			Handle(err)

			lastHash, err = item.ValueCopy(nil)

			return err
		}
	})

	Handle(err)

	return &Blockchain{lastHash, db}
}

func (chain *Blockchain) AddBlock(data string) {

	var lasthash []byte

	err := chain.Database.View(func(txn *badger.Txn) error {

		item, err := txn.Get([]byte("lh"))

		Handle(err)

		lasthash, err = item.ValueCopy(nil)

		return err

	})

	Handle(err)

	newBlock := CreateBlock(data, lasthash)

	err = chain.Database.Update(func(txn *badger.Txn) error {

		err := txn.Set(newBlock.Hash, newBlock.Serialize())

		Handle(err)

		err = txn.Set([]byte("lh"), newBlock.Hash)

		chain.LastHash = newBlock.Hash

		return err
	})

	Handle(err)
}

func (chain *Blockchain) Iterator() *BlockchainIterator {

	iter := &BlockchainIterator{chain.LastHash, chain.Database}

	return iter
}

func (iter *BlockchainIterator) Next() *Block {

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
