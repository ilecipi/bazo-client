package main

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/bazo-blockchain/bazo-client/cli"
	"github.com/bazo-blockchain/bazo-client/client"
	"github.com/bazo-blockchain/bazo-client/cstorage"
	"github.com/bazo-blockchain/bazo-client/network"
	"github.com/bazo-blockchain/bazo-client/util"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	cli2 "github.com/urfave/cli"
	"golang.org/x/crypto/ed25519"
	"net"
	"os"
)
var mainPublicKey = [PUB_KEY_LEN]byte{}
var isPubKey = false;

type PublicKey struct {
	Pk   []int      `json:"Pk"`
}

const (
	PUB_KEY_LEN   = 32
	SIGNATURE_LEN = 64
	HASH_LEN = 32
)

type Transaction struct {
	WalletPubKey   []int      `json:"WalletPubKey"`
	PublicKey   []int      `json:"PublicKey"`
	TxCnt   []int      `json:"TxCnt"`
	TxFee   []int      `json:"TxFee"`
	Header   []int      `json:"Header"`
	Data   []int      `json:"Data"`
	Signature   []int      `json:"Signature"`
}



func main() {
	p2p.InitLogging()
	client.InitLogging()
	logger := util.InitLogger()
	util.Config = util.LoadConfiguration()

	network.Init()
	cstorage.Init("client.db")

	app := cli2.NewApp()

	app.Name = "bazo-client"
	app.Usage = "the command line interface for interacting with the Bazo blockchain implemented in Go."
	app.Version = "1.0.0"
	app.Commands = []cli2.Command{
		cli.GetAccountCommand(logger),
		cli.GetFundsCommand(logger),
		cli.GetNetworkCommand(logger),
		cli.GetRestCommand(),
		cli.GetStakingCommand(logger),
	}
	//TODO move away from here
	go udpStart();
	err := app.Run(os.Args)
	if err != nil {
		logger.Fatal(err)
	}
}

func udpStart() {
	fmt.Println("LISTENING TO UDP")
	ServerConn, _ := net.ListenUDP("udp", &net.UDPAddr{IP: []byte{0, 0, 0, 0}, Port: 5000, Zone: ""})
	defer ServerConn.Close()
	buf := make([]byte, 1024)
	for {
		n, addr, _ := ServerConn.ReadFromUDP(buf)
		var transaction Transaction;
		var pk PublicKey;


		err := json.Unmarshal(buf[0:n], &transaction)
		if (err != nil) {
			fmt.Println("Not a Transaction", err);
		}

		err = json.Unmarshal(buf[0:n], &pk)
		if (err != nil) {
			fmt.Println("Not a publicKey", err);
		}

		if(pk.Pk!=nil){
			for index := range pk.Pk {
				mainPublicKey[index] = byte(pk.Pk[index])
				isPubKey = true;
			}
			fmt.Println("HEX PUB KEY",hex.EncodeToString(mainPublicKey[:]))
		}

		if(transaction.WalletPubKey !=nil){
			publicKey := [PUB_KEY_LEN]byte{}
			for index := range transaction.PublicKey {
				publicKey[index] = byte(transaction.PublicKey[index])
			}

			data := make([]byte, len(transaction.Data))
			for index := range transaction.Data {
				data[index] = byte(transaction.Data[index])
			}

			signature := [SIGNATURE_LEN]byte{}
			for index := range transaction.Signature {
				signature[index] = byte(transaction.Signature[index])
			}
			walletPublicKey := [PUB_KEY_LEN]byte{}

			for index := range transaction.WalletPubKey {
				walletPublicKey[index] = byte(transaction.WalletPubKey[index])
			}

			TxFee := [8]byte{}
			for index := range transaction.TxFee {
				TxFee[8-index-1] = byte(transaction.TxFee[index])
			}
			TxFee64,_ := binary.Uvarint(TxFee[:]);

			TxCnt := [4]byte{}
			for index := range transaction.TxCnt {
				TxCnt[4-index-1] = byte(transaction.TxCnt[index])
			}
			TxCnt32,_ := binary.Uvarint(TxCnt[:]);


			IotTx := protocol.IotTx{
				Header: byte(transaction.Header[0]),
				TxCnt:  uint32(TxCnt32),
				From:   protocol.SerializeHashContent(publicKey),
				To:     protocol.SerializeHashContent(walletPublicKey),
				Sig:    signature,
				Data:   data,
				Fee:    TxFee64,
			}

			txHash := IotTx.Hash()

			isValid := ed25519.Verify(crypto.GetPubKeyFromAddressED(publicKey), txHash[:], signature[:])
			fmt.Println("IS VALID", isValid);
			fmt.Println(IotTx);
			err = network.SendIotTx(util.Config.BootstrapIpport, &IotTx, p2p.IOTTX_BRDCST)



		}

		fmt.Println("Received ", string(buf[0:n]), " from ", addr)
	}
}
