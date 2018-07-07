package client

import (
	"github.com/bazo-blockchain/bazo-client/util"
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
	p2p.InitLogging()
	logger = util.InitLogger()
}

func ProcessTx(args []string) {
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

func ProcessState(fileName string) {
	pubKey, _, err := storage.ExtractKeyFromFile(fileName)
	if err != nil {
		logger.Printf("%v\n%v", err, USAGE_MSG)
		return
	}

	update()

	address := storage.GetAddressFromPubKey(&pubKey)

	logger.Printf("My address: %x\n", address)

	acc, _, err := GetAccount(address)
	if err != nil {
		logger.Println(err)
	} else {
		logger.Printf(acc.String())
	}
}

func Sync() {
	sync()
}
