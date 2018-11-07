package cli

import (
	"github.com/bazo-blockchain/bazo-client/REST"
	"github.com/bazo-blockchain/bazo-client/client"
	"github.com/bazo-blockchain/bazo-client/cstorage"
	"github.com/bazo-blockchain/bazo-client/network"
	"github.com/urfave/cli"
)

func AddStartRestCommand(app *cli.App) {
	command := cli.Command {
		Name:	"startRestService",
		Usage:	"start the REST service",
		Action:	func(c *cli.Context) error {
			//For querying an account state or starting the REST service, the client must establish a connection to the Bazo network.
			network.Init()
			cstorage.Init("client.db")
			client.Sync()
			REST.Init()
			return nil
		},
	}

	app.Commands = append(app.Commands, command)
}