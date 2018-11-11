package account

import (
	"fmt"
	"github.com/bazo-blockchain/bazo-client/network"
	"github.com/bazo-blockchain/bazo-client/util"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/urfave/cli"
	"log"
)

var (
	headerFlag = cli.IntFlag {
		Name: 	"header",
		Usage: 	"header flag",
		Value:	0,
	}

	feeFlag = cli.IntFlag {
		Name: 	"fee",
		Usage:	"specify the fee",
		Value:	1,
	}

	rootkeyFlag = cli.StringFlag {
		Name: 	"rootkey",
		Usage: 	"load root's private key from `FILE`",
	}
)

func GetAccountCommand(logger *log.Logger) cli.Command {
	return cli.Command {
		Name:	"account",
		Usage:	"account management",
		Subcommands: []cli.Command {
			getCheckAccountCommand(logger),
			getCreateAccountCommand(logger),
			getAddAccountCommand(logger),
		},
	}
}



func sendAccountTx(tx protocol.Transaction, logger *log.Logger) error {
	fmt.Printf("chash: %x\n", tx.Hash())

	if err := network.SendTx(util.Config.BootstrapIpport, tx, p2p.ACCTX_BRDCST); err != nil {
		logger.Printf("%v\n", err)
		return err
	} else {
		logger.Printf("Transaction successfully sent to network:\nTxHash: %x%v", tx.Hash(), tx)
	}

	return nil
}