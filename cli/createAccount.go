package cli

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-client/network"
	"github.com/bazo-blockchain/bazo-client/util"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/urfave/cli"
	"log"
	"math/big"
	"os"
)

type createAccountArgs struct {
	header			int
	fee				int
	rootKeyFileName	string
	accountAddress	string
	accountFile		string
}

func AddCreateAccountCommand(app *cli.App, logger *log.Logger) {
	command := cli.Command {
		Name:	"createAccount",
		Usage:	"create a new account",
		Action:	func(c *cli.Context) error {
			args := &createAccountArgs {
				header: 			c.Int("header"),
				fee: 				c.Int("fee"),
				rootKeyFileName: 	c.String("rootkey"),
				accountAddress: 	c.String("accountaddress"),
				accountFile: 		c.String("accountfile"),
			}

			err := args.ValidateInput()
			if err != nil {
				return err
			}

			return createAccount(args, logger)
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
			},
			cli.StringFlag {
				Name: 	"rootkey",
				Usage: 	"load root's public key from `FILE`",
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
		},
	}

	app.Commands = append(app.Commands, command)
}

func createAccount(args *createAccountArgs, logger *log.Logger) error {
	privKey, err := crypto.ExtractECDSAKeyFromFile(args.rootKeyFileName)
	if err != nil {
		return err
	}

	var tx protocol.Transaction
	if len(args.accountAddress) == 128 {
		var newAddress [64]byte
		newPubInt, _ := new(big.Int).SetString(args.accountAddress, 16)
		copy(newAddress[:], newPubInt.Bytes())

		tx, _, err = protocol.ConstrAccTx(byte(args.header), uint64(args.fee), newAddress, privKey, nil, nil)
		if err != nil {
			return err
		}
	} else {
		var newKey *ecdsa.PrivateKey
		//Write the public key to the given textfile
		file, err := os.Create(args.accountFile)
		if err != nil {
			return err
		}

		tx, newKey, err = protocol.ConstrAccTx(byte(args.header), uint64(args.fee), [64]byte{}, privKey, nil, nil)
		if err != nil {
			return err
		}

		_, err = file.WriteString(string(newKey.X.Text(16)) + "\n")
		_, err = file.WriteString(string(newKey.Y.Text(16)) + "\n")
		_, err = file.WriteString(string(newKey.D.Text(16)) + "\n")

		if err != nil {
			return errors.New(fmt.Sprintf("failed to write key to file %v", args.accountFile))
		}
	}

	fmt.Printf("chash: %x\n", tx.Hash())

	if err != nil {
		logger.Printf("%v\n", err)
		return err
	}

	if err := network.SendTx(util.Config.BootstrapIpport, tx, p2p.ACCTX_BRDCST); err != nil {
		logger.Printf("%v\n", err)
		return err
	} else {
		logger.Printf("Transaction successfully sent to network:\nTxHash: %x%v", tx.Hash(), tx)
	}

	return nil
}

func (args createAccountArgs) ValidateInput() error {
	if args.fee <= 0 {
		return errors.New("invalid argument: fee must be > 0")
	}

	if len(args.rootKeyFileName) == 0 {
		return errors.New("argument missing: rootKeyFileName")
	}

	if len(args.accountAddress) == 0 && len(args.accountFile) == 0 {
		return errors.New("argument missing: accountAddress or accountFile")
	}

	if len(args.accountFile) == 0 && len(args.accountAddress) != 128 {
		return errors.New("invalid argument: accountAddress")
	}

	return nil
}
