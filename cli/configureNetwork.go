package cli

import (
	"errors"
	"github.com/bazo-blockchain/bazo-client/network"
	"github.com/bazo-blockchain/bazo-client/util"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/urfave/cli"
	"log"
)

type configureNetworkArgs struct {
	header			int
	fee				int
	txcnt			int
	rootKeyFileName	string
	optionId		uint8
	payload			uint64
}

type configOption struct {
	id				uint8
	name			string
	usage		string
}

func AddConfigureNetworkCommand(app *cli.App, logger *log.Logger) {
	options := []configOption {
		{ id: 1, name: "setBlockSize", usage: "set the size of blocks (in bytes)" },
		{ id: 2, name: "setDifficultyInterval", usage: "set the difficulty interval (in number of blocks)" },
		{ id: 3, name: "setMinimumFee", usage: "set the minimum fee (in Bazo coins)" },
		{ id: 4, name: "setBlockInterval", usage: "set the block interval (in seconds)" },
		{ id: 5, name: "setBlockReward", usage: "set the block reward (in Bazo coins)" },
	}

	command := cli.Command {
		Name:	"configureNetwork",
		Usage:	"configure the network",
		Action:	func(c *cli.Context) error {
			for _, option := range options {
				if !c.IsSet(option.name) { continue }

				args := &configureNetworkArgs {
					header: 			c.Int("header"),
					fee: 				c.Int("fee"),
					txcnt: 				c.Int("txcnt"),
					rootKeyFileName: 	c.String("rootkey"),
					optionId:			option.id,
					payload:			c.Uint64(option.name),
				}

				err := args.ValidateInput()
				if err != nil {
					return err
				}

				err = configureNetwork(args, logger)
				if err != nil {
					return err
				}
			}

			return nil
		},
		Flags:	[]cli.Flag {
			cli.IntFlag {
				Name: 	"header",
				Usage: 	"header flag",
				Value:	0,
			},
			cli.IntFlag {
				Name: 	"fee",
				Usage:	"specify the fee",
				Value: 	1,
			},
			cli.IntFlag {
				Name: 	"txcount",
				Usage:	"the sender's current transaction counter",
			},
			cli.StringFlag {
				Name: 	"rootkey",
				Usage: 	"load root's public key from `FILE`",
				Value: 	"key.txt",
			},
		},
	}

	for _, option := range options {
		flag := cli.Uint64Flag { Name: option.name, Usage: option.usage }
		command.Flags = append(command.Flags, flag)
	}

	app.Commands = append(app.Commands, command)
}

func configureNetwork(args *configureNetworkArgs, logger *log.Logger) error {
	privKey, err := crypto.ExtractECDSAKeyFromFile(args.rootKeyFileName)
	if err != nil {
		return err
	}

	tx, err := protocol.ConstrConfigTx(byte(args.header), uint8(args.optionId), uint64(args.payload), uint64(args.fee), uint8(args.txcnt), privKey)
	if err != nil {
		return err
	}

	if tx == nil {
		return errors.New("transaction encoding failed")
	}

	if err := network.SendTx(util.Config.BootstrapIpport, tx, p2p.CONFIGTX_BRDCST); err != nil {
		logger.Printf("%v\n", err)
		return err
	} else {
		logger.Printf("Transaction successfully sent to network:\nTxHash: %x%v", tx.Hash(), tx)
	}

	return nil
}

func (args configureNetworkArgs) ValidateInput() error {
	if args.fee <= 0 {
		return errors.New("invalid argument: fee must be > 0")
	}

	if args.txcnt < 0 {
		return errors.New("invalid argument: txcnt must be >= 0")
	}

	if len(args.rootKeyFileName) == 0 {
		return errors.New("argument missing: rootKeyFileName")
	}

	return nil
}
