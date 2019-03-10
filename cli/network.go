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

type networkArgs struct {
	header      	int
	fee         	uint64
	txcount     	int
	rootWalletFile 	string
	optionId    	uint8
	payload     	uint64
}

type configOption struct {
	id				uint8
	name			string
	usage			string
}

func GetNetworkCommand(logger *log.Logger) cli.Command {
	options := []configOption {
		{ id: 1, name: "setBlockSize", usage: "set the size of blocks (in bytes)" },
		{ id: 2, name: "setDifficultyInterval", usage: "set the difficulty interval (in number of blocks)" },
		{ id: 3, name: "setMinimumFee", usage: "set the minimum fee (in Bazo coins)" },
		{ id: 4, name: "setBlockInterval", usage: "set the block interval (in seconds)" },
		{ id: 5, name: "setBlockReward", usage: "set the block reward (in Bazo coins)" },
	}

	command := cli.Command {
		Name:	"network",
		Usage:	"configure the network",
		Action:	func(c *cli.Context) error {
			optionsSetByUser := 0
			for _, option := range options {
				if !c.IsSet(option.name) { continue }

				optionsSetByUser++

				args := &networkArgs {
					header:      	c.Int("header"),
					fee:         	c.Uint64("fee"),
					rootWalletFile: c.String("rootwallet"),
					optionId:    	option.id,
					payload:     	c.Uint64(option.name),
					txcount:		c.Int("txcount"),
				}

				err := configureNetwork(args, logger)
				if err != nil {
					return err
				}
			}

			if optionsSetByUser == 0 {
				return errors.New("specify at least one configuration option")
			}

			return nil
		},
		Flags: []cli.Flag {
			cli.IntFlag {
				Name: 	"header",
				Usage: 	"header flag",
				Value:	0,
			},
			cli.Uint64Flag {
				Name: 	"fee",
				Usage:	"specify the fee",
				Value: 	1,
			},
			cli.IntFlag {
				Name: 	"txcount",
				Usage:	"the sender's current transaction counter",
			},
			cli.StringFlag {
				Name: 	"rootwallet",
				Usage: 	"load root's public key from `FILE`",
			},
		},
	}

	for _, option := range options {
		flag := cli.Uint64Flag { Name: option.name, Usage: option.usage }
		command.Flags = append(command.Flags, flag)
	}

	return command
}

func configureNetwork(args *networkArgs, logger *log.Logger) error {
	err := args.ValidateInput()
	if err != nil {
		return err
	}

	privKey, err := crypto.ExtractEDPrivKeyFromFile(args.rootWalletFile)
	if err != nil {
		return err
	}

	tx, err := protocol.ConstrConfigTx(
		byte(args.header),
		uint8(args.optionId),
		uint64(args.payload),
		uint64(args.fee),
		uint8(args.txcount),
		privKey)

	if err != nil {
		return err
	}

	if tx == nil {
		return errors.New("transaction encoding failed")
	}

	if err := network.SendTx(util.Config.BootstrapIpport, tx, p2p.CONFIGTX_BRDCST); err != nil {
		//logger.Printf("%v\n", err)
		return err
	} else {
		//logger.Printf("Transaction successfully sent to network:\nTxHash: %x%v", tx.Hash(), tx)
	}

	return nil
}

func (args networkArgs) ValidateInput() error {
	if args.fee <= 0 {
		return errors.New("invalid argument: fee must be > 0")
	}

	if args.txcount < 0 {
		return errors.New("invalid argument: txcnt must be >= 0")
	}

	if len(args.rootWalletFile) == 0 {
		return errors.New("argument missing: rootwallet")
	}

	return nil
}
