package client

import (
	"github.com/bazo-blockchain/bazo-client/util"
	"log"
)

var (
	logger     *log.Logger
)

func InitLogging() {
	logger = util.InitLogger()
}

func put(slice []*FundsTxJson, tx *FundsTxJson) {
	for i := 0; i < 9; i++ {
		slice[i] = slice[i+1]
	}

	slice[9] = tx
}
