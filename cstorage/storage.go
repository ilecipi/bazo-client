package cstorage

import (
	"fmt"
	"github.com/bazo-blockchain/bazo-client/util"
	"github.com/boltdb/bolt"
	"log"
	"time"
)

var (
	db     *bolt.DB
	logger *log.Logger
)

const (
	ERROR_MSG = "Initiate storage aborted: "
)

//Entry function for the storage package
func Init(dbname string) {
	logger = util.InitLogger()

	var err error
	db, err = bolt.Open(dbname, 0600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		logger.Fatal(ERROR_MSG, err)
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucket([]byte("blockheaders"))
		if err != nil {
			return fmt.Errorf(ERROR_MSG+"Create bucket: %s", err)
		}

		return nil
	})

	db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucket([]byte("lastblockheader"))
		if err != nil {
			return fmt.Errorf(ERROR_MSG+"Create bucket: %s", err)
		}

		return nil
	})
}

func TearDown() {
	db.Close()
}
