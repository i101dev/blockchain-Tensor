package blockchain

import (
	"bytes"
	"encoding/hex"
	"log"

	"github.com/dgraph-io/badger"
	"github.com/i101dev/blockchain-Tensor/util"
)

var (
	utxoPrefix   = []byte("utxo-")
	prefixLength = len(utxoPrefix)
)

type UTXOSet struct {
	Blockchain *Blockchain
}

func (utxo *UTXOSet) DeleteByPrefix(prefix []byte) {
	//
	// -------------------------------------------------------------
	deleteKeys := func(keysForDelete [][]byte) error {
		if err := utxo.Blockchain.Database.Update(func(txn *badger.Txn) error {
			for _, key := range keysForDelete {
				if err := txn.Delete(key); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}
	// -------------------------------------------------------------
	//

	collectSize := 100000
	utxo.Blockchain.Database.View(func(txn *badger.Txn) error {

		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false

		it := txn.NewIterator(opts)
		defer it.Close()

		keysForDelete := make([][]byte, 0, collectSize)
		keysCollected := 0

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {

			key := it.Item().KeyCopy(nil)
			keysForDelete = append(keysForDelete, key)

			keysCollected++
			if keysCollected == collectSize {

				if err := deleteKeys(keysForDelete); err != nil {
					log.Panic(err)
				}

				keysForDelete = make([][]byte, 0, collectSize)
				keysCollected = 0
			}
		}

		if keysCollected > 0 {
			if err := deleteKeys(keysForDelete); err != nil {
				log.Panic(err)
			}
		}

		return nil
	})
}

func (utxo UTXOSet) CountTransactions() int {
	db := utxo.Blockchain.Database
	counter := 0

	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions

		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			counter++
		}

		return nil
	})

	util.HandleError(err, "CountTransactions")

	return counter
}

func (utxo *UTXOSet) Update(block *Block) {
	db := utxo.Blockchain.Database

	err := db.Update(func(txn *badger.Txn) error {
		for _, tx := range block.Transactions {
			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					updatedOuts := TxOutputs{}
					inID := append(utxoPrefix, in.ID...)
					item, err := txn.Get(inID)
					util.HandleError(err, "Update 1")
					v, err := item.ValueCopy(nil)
					util.HandleError(err, "Update 2")

					outs := DeserializeTxOutputs(v)

					for outIdx, out := range outs.Outputs {
						//
						// each input contains a reference to the output it came from
						//
						if outIdx != in.Out {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}

					if len(updatedOuts.Outputs) == 0 {
						if err := txn.Delete(inID); err != nil {
							log.Panic(err)
						}

					} else {
						if err := txn.Set(inID, updatedOuts.Serialize()); err != nil {
							log.Panic(err)
						}
					}
				}
			}

			newOutputs := TxOutputs{}
			for _, out := range tx.Outputs {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}

			txID := append(utxoPrefix, tx.HashID...)
			if err := txn.Set(txID, newOutputs.Serialize()); err != nil {
				log.Panic(err)
			}
		}

		return nil
	})

	util.HandleError(err, "Update 3")
}

func (utxo UTXOSet) Reindex() {
	db := utxo.Blockchain.Database

	utxo.DeleteByPrefix(utxoPrefix)

	UTXO := utxo.Blockchain.FindUTXO()

	err := db.Update(func(txn *badger.Txn) error {

		for txId, outs := range UTXO {

			key, err := hex.DecodeString(txId)

			if err != nil {
				return err
			}

			key = append(utxoPrefix, key...)

			err = txn.Set(key, outs.Serialize())
			util.HandleError(err, "Reindex 1")
		}

		return nil
	})

	util.HandleError(err, "Reindex 2")
}

func (u UTXOSet) FindUnspentTransactions(pubKeyHash []byte) []TxOutput {
	var UTXOs []TxOutput

	db := u.Blockchain.Database

	err := db.View(func(txn *badger.Txn) error {

		opts := badger.DefaultIteratorOptions

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {

			item := it.Item()
			v, err := item.ValueCopy(nil)
			util.HandleError(err, "FindUnspentTransactions 1")

			outs := DeserializeTxOutputs(v)

			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}

		}
		return nil
	})

	util.HandleError(err, "FindUnspentTransactions 2")

	return UTXOs
}

func (u UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {

	unspentOuts := make(map[string][]int)
	accumulated := 0
	db := u.Blockchain.Database

	err := db.View(func(txn *badger.Txn) error {

		opts := badger.DefaultIteratorOptions

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {

			item := it.Item()
			k := item.Key()
			v, err := item.ValueCopy(nil)

			util.HandleError(err, "FindSpendableOutputs 1")

			k = bytes.TrimPrefix(k, utxoPrefix)
			txID := hex.EncodeToString(k)
			outs := DeserializeTxOutputs(v)

			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOuts[txID] = append(unspentOuts[txID], outIdx)
				}
			}
		}
		return nil
	})

	util.HandleError(err, "FindSpendableOutputs 2")

	return accumulated, unspentOuts
}
