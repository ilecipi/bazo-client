package main

import (
	"github.com/bazo-blockchain/bazo-client/REST"
	"github.com/bazo-blockchain/bazo-client/client"
	"github.com/bazo-blockchain/bazo-client/network"
	"os"
)

func main() {
	client.Init()

	if len(os.Args) > 1 {
		if os.Args[1] == "accTx" || os.Args[1] == "fundsTx" || os.Args[1] == "configTx" || os.Args[1] == "stakeTx" {
			client.ProcessTx(os.Args[1:])
		}

		return
	}

	//For querying an account state or starting the REST service, the client must establish a connection to the Bazo network.
	network.Init()

	if len(os.Args) == 2 {
		client.ProcessState(os.Args[1])

		return
	}

	client.Sync()
	REST.Init()
}
