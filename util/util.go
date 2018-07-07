package util

import (
	"github.com/bazo-blockchain/bazo-miner/storage"
	"log"
	"os"
)

const (
	LIGHT_CLIENT_SERVER_IP   = storage.BOOTSTRAP_SERVER_IP
	LIGHT_CLIENT_SERVER_PORT = ":8001"
	LIGHT_CLIENT_SERVER      = LIGHT_CLIENT_SERVER_IP + LIGHT_CLIENT_SERVER_PORT

	MULTISIG_SERVER_IP   = storage.BOOTSTRAP_SERVER_IP
	MULTISIG_SERVER_PORT = ":8002"
	MULTISIG_SERVER      = MULTISIG_SERVER_IP + MULTISIG_SERVER_PORT
)

func InitLogger() *log.Logger {
	return log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
}
