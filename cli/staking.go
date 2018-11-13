package cli

import (
	"crypto/rsa"
	"errors"
	"github.com/bazo-blockchain/bazo-client/network"
	"github.com/bazo-blockchain/bazo-client/util"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/urfave/cli"
	"log"
)

type stakingArgs struct {
	header			int
	fee				uint64
	walletFile		string
	commitment		string
	stakingValue	bool
}

func GetStakingCommand(logger *log.Logger) cli.Command {
	headerFlag := cli.IntFlag {
		Name: 	"header",
		Usage: 	"header flag",
		Value:	0,
	}

	feeFlag := cli.Uint64Flag {
		Name: 	"fee",
		Usage:	"specify the fee",
		Value: 	1,
	}

	walletFlag := cli.StringFlag {
		Name: 	"wallet, w",
		Usage: 	"load validator's public key from `FILE`",
		Value: 	"wallet.txt",
	}

	return cli.Command {
		Name:	"staking",
		Usage:	"enable or disable staking",
		Subcommands: []cli.Command {
			{
				Name: "enable",
				Usage: "join the pool of validators",
				Action:	func(c *cli.Context) error {
					args := parseStakingArgs(c)
					args.stakingValue = true
					return toggleStaking(args, logger)
				},
				Flags: []cli.Flag {
					headerFlag,
					feeFlag,
					walletFlag,
					cli.StringFlag {
						Name: 	"commitment",
						Usage: 	"load valiadator's commitment key from `FILE`",
						Value: 	"commitment.txt",
					},
				},
			},
			{
				Name: "disable",
				Usage: "leave the pool of validators",
				Action:	func(c *cli.Context) error {
					args := parseStakingArgs(c)
					args.stakingValue = false
					return toggleStaking(args, logger)
				},
				Flags: []cli.Flag {
					headerFlag,
					feeFlag,
					walletFlag,
				},
			},
		},
	}
}

func parseStakingArgs(c *cli.Context) *stakingArgs {
	return &stakingArgs {
		header: 			c.Int("header"),
		fee: 				c.Uint64("fee"),
		walletFile:	 		c.String("wallet"),
		commitment:			c.String("commitment"),
	}
}

func toggleStaking(args *stakingArgs, logger *log.Logger) error {
	err := args.ValidateInput()
	if err != nil {
		return err
	}

	privKey, err := crypto.ExtractECDSAKeyFromFile(args.walletFile)
	if err != nil {
		return err
	}

	accountPubKey := crypto.GetAddressFromPubKey(&privKey.PublicKey)

	commPubKey := &rsa.PublicKey{}
	if args.stakingValue {
		commPrivKey, err := crypto.ExtractRSAKeyFromFile(args.commitment)
		if err != nil {
			return err
		}
		commPubKey = &commPrivKey.PublicKey
	}

	tx, err := protocol.ConstrStakeTx(
		byte(args.header),
		uint64(args.fee),
		args.stakingValue,
		protocol.SerializeHashContent(accountPubKey),
		privKey,
		commPubKey,
	)

	if err != nil {
		return err
	}

	if tx == nil {
		return errors.New("transaction encoding failed")
	}

	if err := network.SendTx(util.Config.BootstrapIpport, tx, p2p.STAKETX_BRDCST); err != nil {
		logger.Printf("%v\n", err)
		return err
	} else {
		logger.Printf("Transaction successfully sent to network:\nTxHash: %x%v", tx.Hash(), tx)
	}

	return nil
}

func (args stakingArgs) ValidateInput() error {
	if args.fee <= 0 {
		return errors.New("invalid argument: fee must be > 0")
	}

	if len(args.walletFile) == 0 {
		return errors.New("argument missing: wallet")
	}

	if args.stakingValue && len(args.commitment) == 0 {
		return errors.New("argument missing: commitment")
	}

	return nil
}
