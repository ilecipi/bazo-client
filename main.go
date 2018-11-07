package main

import (
	"github.com/bazo-blockchain/bazo-client/cli"
	"github.com/bazo-blockchain/bazo-client/client"
	"github.com/bazo-blockchain/bazo-client/util"
	cli2 "github.com/urfave/cli"
	"os"
)

func main() {
	logger := util.InitLogger()

	client.Init()

	app := cli2.NewApp()

	// Global app config
	app.Name = "bazo-client"
	app.Usage = "the command line interface for interacting with the Bazo blockchain implemented in Go."
	app.Version = "1.0.0"
	app.EnableBashCompletion = true

	cli.AddSendFundsCommand(app, logger)
	cli.AddCreateAccountCommand(app, logger)
	cli.AddCheckAccountCommand(app, logger)
	cli.AddStartRestCommand(app)

	err := app.Run(os.Args)
	if err != nil {
		logger.Fatal(err)
	}
}
