package main

import (
	"github.com/bazo-blockchain/bazo-client/cli/account"
	"github.com/bazo-blockchain/bazo-client/cli/funds"
	"github.com/bazo-blockchain/bazo-client/cli/network"
	"github.com/bazo-blockchain/bazo-client/cli/rest"
	"github.com/bazo-blockchain/bazo-client/cli/staking"
	"github.com/bazo-blockchain/bazo-client/client"
	"github.com/bazo-blockchain/bazo-client/util"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	cli2 "github.com/urfave/cli"
	"os"
)

func main() {
	p2p.InitLogging()
	client.InitLogging()
	logger := util.InitLogger()
	util.Config = util.LoadConfiguration()

	app := cli2.NewApp()

	app.Name = "bazo-client"
	app.Usage = "the command line interface for interacting with the Bazo blockchain implemented in Go."
	app.Version = "1.0.0"
	app.Commands = []cli2.Command {
		account.GetAccountCommand(logger),
		funds.GetFundsCommand(logger),
		network.GetNetworkCommand(logger),
		rest.GetRestCommand(),
		staking.GetStakingCommand(logger),
	}

	err := app.Run(os.Args)
	if err != nil {
		logger.Fatal(err)
	}
}
