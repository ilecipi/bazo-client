package cstorage

import (
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/boltdb/bolt"
)

func ReadBlockHeader(hash [32]byte) (header *protocol.Block) {
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blockheaders"))
		encodedHeader := b.Get(hash[:])
		header = header.DecodeHeader(encodedHeader)

		return nil
	})

	if header == nil {
		return nil
	}

	return header
}

func ReadLastBlockHeader() (header *protocol.Block) {
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("lastblockheader"))
		cb := b.Cursor()
		_, encodedHeader := cb.First()
		header = header.DecodeHeader(encodedHeader)

		return nil
	})

	if header == nil {
		return nil
	}

	return header
}
