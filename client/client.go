package client

import (
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"log"
	"os"
)

var (
	err     error
	msgType uint8
	tx      protocol.Transaction
	logger  *log.Logger
)

const (
	USAGE_MSG = "Usage: bazo-client [pubKey|accTx|fundsTx|configTx|stakeTx] ...\n"
)

func Init() {
	logger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func State(keyFile string) {
	pubKey, _, err := storage.ExtractKeyFromFile(keyFile)
	if err != nil {
		logger.Printf("%v\n%v", err, USAGE_MSG)
		return
	}

	InitState()

	accAddress := storage.GetAddressFromPubKey(&pubKey)

	logger.Printf("My address: %x\n", accAddress)

	acc, _, err := GetAccount(accAddress)
	if err != nil {
		logger.Println(err)
	} else {
		logger.Printf(acc.String())
	}
}

func Process(args []string) {
	switch args[0] {
	case "accTx":
		tx, err = parseAccTx(os.Args[2:])
		msgType = p2p.ACCTX_BRDCST
	case "fundsTx":
		tx, err = parseFundsTx(os.Args[2:])
		msgType = p2p.FUNDSTX_BRDCST
	case "configTx":
		tx, err = parseConfigTx(os.Args[2:])
		msgType = p2p.CONFIGTX_BRDCST
	case "stakeTx":
		tx, err = parseStakeTx(os.Args[2:])
		msgType = p2p.STAKETX_BRDCST
	}
	if err != nil {
		logger.Printf("%v\n", err)
		return
	}

	if err := SendTx(storage.BOOTSTRAP_SERVER, tx, msgType); err != nil {
		logger.Printf("%v\n", err)
	} else {
		logger.Printf("Transaction successfully sent to network:%v", tx)
	}
}
