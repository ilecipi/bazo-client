package account

import (
	"errors"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/urfave/cli"
	"log"
	"math/big"
)

type addAccountArgs struct {
	header			int
	fee				int
	rootkeyFile		string
	address	string
}

func getAddAccountCommand(logger *log.Logger) cli.Command {
	return cli.Command {
		Name: "add",
			Usage: "add an existing account",
			Action: func(c *cli.Context) error {
			args := &addAccountArgs {
				header: 		c.Int("header"),
				fee: 			c.Int("fee"),
				rootkeyFile: 	c.String("rootkey"),
				address: c.String("address"),
			}

			return addAccount(args, logger)
		},
			Flags: []cli.Flag {
			headerFlag,
			feeFlag,
			rootkeyFlag,
			cli.StringFlag {
				Name: 	"address",
				Usage: 	"the account's address",
			},
		},
	}
}

func addAccount(args *addAccountArgs, logger *log.Logger) error {
	err := args.ValidateInput()
	if err != nil {
		return err
	}

	privKey, err := crypto.ExtractECDSAKeyFromFile(args.rootkeyFile)
	if err != nil {
		return err
	}

	var newAddress [64]byte
	newPubInt, _ := new(big.Int).SetString(args.address, 16)
	copy(newAddress[:], newPubInt.Bytes())

	tx, _, err := protocol.ConstrAccTx(byte(args.header), uint64(args.fee), newAddress, privKey, nil, nil)
	if err != nil {
		return err
	}

	return sendAccountTx(tx, logger)
}

func (args addAccountArgs) ValidateInput() error {
	if args.fee <= 0 {
		return errors.New("invalid argument: fee must be > 0")
	}

	if len(args.rootkeyFile) == 0 {
		return errors.New("argument missing: rootkeyFile")
	}

	if len(args.address) == 0 {
		return errors.New("argument missing: accountAddress")
	}

	if len(args.address) != 128 {
		return errors.New("invalid argument length: accountAddress")
	}

	return nil
}