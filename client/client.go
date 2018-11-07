package client

import (
	"github.com/bazo-blockchain/bazo-client/network"
	"github.com/bazo-blockchain/bazo-client/util"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"log"
	"os"
)

var (
	err     error
	msgType uint8
	tx      protocol.Transaction
	logger  *log.Logger
)

func Init() {
	p2p.InitLogging()
	logger = util.InitLogger()
	util.Config = util.LoadConfiguration()
}

func ProcessTx(args []string) {
	switch args[0] {
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

	if err := network.SendTx(util.Config.BootstrapIpport, tx, msgType); err != nil {
		logger.Printf("%v\n", err)
	} else {
		logger.Printf("Transaction successfully sent to network:\nTxHash: %x%v", tx.Hash(), tx)
	}
}
