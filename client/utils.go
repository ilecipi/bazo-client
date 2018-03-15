package client

import (
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/boltdb/bolt"
)

func put(slice []*FundsTxJson, tx *FundsTxJson) {
	for i := 0; i < 9; i++ {
		slice[i] = slice[i+1]
	}

	slice[9] = tx
}

func WriteLastBlockHeader(header *protocol.Block) (err error) {

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("lastblockheader"))
		err := b.Put(header.Hash[:], header.Encode())
		return err
	})

	return err
}

func DeleteLastBlockHeader(hash [32]byte) {

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("lastblockheader"))
		err := b.Delete(hash[:])
		return err
	})
}

func ReadLastBlockHeader() (header *protocol.Block) {

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("lastblockheader"))
		cb := b.Cursor()
		_, encodedBlockHeader := cb.First()
		header = header.Decode(encodedBlockHeader)
		return nil
	})

	if header == nil {
		return nil
	}

	return header
}

func WriteBlockHeader(header *protocol.Block) (err error) {

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blockheaders"))
		err := b.Put(header.Hash[:], header.Encode())
		return err
	})

	return err
}

func DeleteBlockHeader(hash [32]byte) {

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blockheaders"))
		err := b.Delete(hash[:])
		return err
	})
}

func ReadBlockHeader(hash [32]byte) (header *protocol.Block) {

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blockheaders"))
		encodedBlock := b.Get(hash[:])
		header = header.Decode(encodedBlock)
		return nil
	})

	if header == nil {
		return nil
	}

	return header
}