package cstorage

import "github.com/boltdb/bolt"

func DeleteBlockHeader(hash [32]byte) {
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blockheaders"))
		err := b.Delete(hash[:])

		return err
	})
}
