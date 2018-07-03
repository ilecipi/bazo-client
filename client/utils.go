package client

import (
	"log"
	"os"
)

func put(slice []*FundsTxJson, tx *FundsTxJson) {
	for i := 0; i < 9; i++ {
		slice[i] = slice[i+1]
	}

	slice[9] = tx
}

func InitLogging() *log.Logger {
	return log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
}