package cli

import (
	"errors"
	"github.com/bazo-blockchain/bazo-client/client"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"github.com/urfave/cli"
	"log"
	"math/big"
)

type checkAccountArgs struct {
	address		string
	walletFile	string
}

func getCheckAccountCommand(logger *log.Logger) cli.Command {
	return cli.Command {
		Name: "check",
		Usage: "check account state",
			Action: func(c *cli.Context) error {
			args := &checkAccountArgs {
				address:	c.String("address"),
				walletFile:	c.String("wallet"),
			}

			return checkAccount(args, logger)
		},
		Flags: []cli.Flag {
			cli.StringFlag {
				Name: 	"address",
				Usage: 	"the account's 128 byte address",
			},
			cli.StringFlag {
				Name: 	"wallet",
				Usage: 	"load the account's 128 byte address from `FILE`",
				Value: 	"wallet.txt",
			},
		},
	}
}

func checkAccount(args *checkAccountArgs, logger *log.Logger) error {
	err := args.ValidateInput()
	if err != nil {
		return err
	}

	var address [64]byte
	if len(args.address) == 128 {
		newPubInt, _ := new(big.Int).SetString(args.address, 16)
		copy(address[:], newPubInt.Bytes())
	} else {
		privKey, err := crypto.ExtractECDSAKeyFromFile(args.walletFile)
		if err != nil {
			logger.Printf("%v\n", err)
			return err
		}

		address = crypto.GetAddressFromPubKey(&privKey.PublicKey)
	}

	logger.Printf("My address: %x\n", address)

	acc, _, err := client.CheckAccount(address)
	if err != nil {
		logger.Println(err)
		return err
	} else {
		logger.Printf(acc.String())
	}

	return nil
}

func (args checkAccountArgs) ValidateInput() error {
	if len(args.address) == 0 && len(args.walletFile) == 0 {
		return errors.New("argument missing: address or wallet")
	}

	if len(args.walletFile) == 0 && len(args.address) != 128 {
		return errors.New("invalid argument: address")
	}

	return nil
}