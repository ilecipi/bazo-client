package cli

import (
	"github.com/bazo-blockchain/bazo-client/client"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"github.com/urfave/cli"
	"log"
)

func AddCheckAccountCommand(app *cli.App, logger *log.Logger) {
	command := cli.Command {
		Name:	"checkAccount",
		Usage:	"check the account's state",
		Action:	func(c *cli.Context) error {
			filename := c.String("filename")

			privKey, err := crypto.ExtractECDSAKeyFromFile(filename)
			if err != nil {
				logger.Printf("%v\n", err)
				return err
			}

			address := crypto.GetAddressFromPubKey(&privKey.PublicKey)
			logger.Printf("My address: %x\n", address)


			acc, _, err := client.CheckAccount(address)
			if err != nil {
				logger.Println(err)
			} else {
				logger.Printf(acc.String())
			}

			return err
		},
		Flags:	[]cli.Flag {
			cli.StringFlag {
				Name: 	"filename",
				Usage: 	"load the account's public key from `FILE`",
			},
		},
	}

	app.Commands = append(app.Commands, command)
}