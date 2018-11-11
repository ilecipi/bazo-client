package rest

import (
	"github.com/bazo-blockchain/bazo-client/REST"
	"github.com/bazo-blockchain/bazo-client/client"
	"github.com/bazo-blockchain/bazo-client/cstorage"
	"github.com/bazo-blockchain/bazo-client/network"
	"github.com/urfave/cli"
)

func GetRestCommand() cli.Command {
	return cli.Command {
		Name:	"rest",
		Usage:	"start the REST service",
		Action:	func(c *cli.Context) error {
			network.Init()
			cstorage.Init("client.db")
			client.Sync()
			REST.Init()
			return nil
		},
	}
}