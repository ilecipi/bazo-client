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
	"math/big"
)

type toggleStakingArgs struct {
	header			int
	fee				int
	keyFile			string
	accountAddress	string
	accountFile		string
	commitmentFile	string
	enable			bool
	disable			bool
}

func AddConfigureStakingCommand(app *cli.App, logger *log.Logger) {
	command := cli.Command {
		Name:	"configureStaking",
		Usage:	"enable or disable staking",
		Action:	func(c *cli.Context) error {
			args := &toggleStakingArgs {
				header: 			c.Int("header"),
				fee: 				c.Int("fee"),
				keyFile:	 		c.String("key"),
				accountAddress: 	c.String("accountaddress"),
				accountFile: 		c.String("accountfile"),
				commitmentFile:		c.String("commitmentfile"),
				enable:				c.Bool("enable"),
				disable:			c.Bool("disable"),
			}

			err := args.ValidateInput()
			if err != nil {
				return err
			}

			return toggleStaking(args, logger)
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
			cli.StringFlag {
				Name: 	"key, k",
				Usage: 	"load validator's public key from `FILE`",
				Value: 	"key.txt",
			},
			cli.StringFlag {
				Name: 	"accountaddress",
				Usage: 	"create account for existing 128 byte account address",
			},
			cli.StringFlag {
				Name: 	"accountfile",
				Usage: 	"save new account's private key to `FILE`",
				Value: 	"account.txt",
			},
			cli.StringFlag {
				Name: 	"commitmentfile",
				Usage: 	"load valiadator's commitment key from `FILE`",
				Value: 	"commitment.txt",
			},
			cli.BoolFlag {
				Name: 	"enable",
				Usage: 	"enable staking",
			},
			cli.BoolFlag {
				Name: 	"disable",
				Usage: 	"disable staking",
			},
		},
	}

	app.Commands = append(app.Commands, command)
}

func toggleStaking(args *toggleStakingArgs, logger *log.Logger) error {
	var accountPubKey [64]byte
	if len(args.accountAddress) == 128 {
		newPubInt, _ := new(big.Int).SetString(args.accountAddress, 16)
		copy(accountPubKey[:], newPubInt.Bytes())
	} else {
		pubKey, err := crypto.ExtractECDSAPublicKeyFromFile(args.accountFile)
		if err != nil {
			return err
		}
		accountPubKey = crypto.GetAddressFromPubKey(pubKey)
	}

	privKey, err := crypto.ExtractECDSAKeyFromFile(args.keyFile)
	if err != nil {
		return err
	}

	var stakingValue bool
	var commPubKey *rsa.PublicKey
	if args.enable {
		commPrivKey, err := crypto.ExtractRSAKeyFromFile(args.commitmentFile)
		if err != nil {
			return err
		}
		commPubKey = &commPrivKey.PublicKey
		stakingValue = true
	} else {
		stakingValue = false
	}

	tx, err := protocol.ConstrStakeTx(
		byte(args.header),
		uint64(args.fee),
		stakingValue,
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


	if err != nil {
		logger.Printf("%v\n", err)
		return err
	}

	if err := network.SendTx(util.Config.BootstrapIpport, tx, p2p.STAKETX_BRDCST); err != nil {
		logger.Printf("%v\n", err)
		return err
	} else {
		logger.Printf("Transaction successfully sent to network:\nTxHash: %x%v", tx.Hash(), tx)
	}

	return nil
}

func (args toggleStakingArgs) ValidateInput() error {
	if args.fee <= 0 {
		return errors.New("invalid argument: fee must be > 0")
	}

	if len(args.keyFile) == 0 {
		return errors.New("argument missing: keyFile")
	}

	if len(args.accountAddress) == 0 && len(args.accountFile) == 0 {
		return errors.New("argument missing: accountAddress or accountFile")
	}

	if len(args.accountFile) == 0 && len(args.accountAddress) != 128 {
		return errors.New("invalid argument: accountAddress")
	}

	if len(args.commitmentFile) == 0 {
		return errors.New("argument missing: commitmentFile")
	}

	if args.enable && args.disable {
		return errors.New("invalid argument: enable and disable specified at the same time")
	}

	if !args.enable && !args.disable {
		return errors.New("missing argument: use --enable or --disable to configure staking")
	}

	return nil
}
