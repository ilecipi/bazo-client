package UDP

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/bazo-blockchain/bazo-client/network"
	"github.com/bazo-blockchain/bazo-client/util"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"log"
	"net"
)

const (
	PUB_KEY_LEN   = 32
	SIGNATURE_LEN = 64
)

var (
	logger *log.Logger
)

type IoTTransaction struct {
	To        []byte `json:"To"`
	From      []byte `json:"From"`
	TxCnt     []byte `json:"TxCnt"`
	TxFee     []byte `json:"TxFee"`
	Header    []byte `json:"Header"`
	Data      []byte `json:"Data"`
	Signature [64]byte	 `json:"Signature"`
}

func Init() {
	logger = util.InitLogger()
	logger.Printf("%v\n\n", "Listening to incoming UDP packets...")
	ServerConn, _ := net.ListenUDP("udp", &net.UDPAddr{IP: []byte{0, 0, 0, 0}, Port: 5001, Zone: ""})

	defer ServerConn.Close()
	buf := make([]byte, 1024)
	for {
		n, addr, _ := ServerConn.ReadFromUDP(buf)
		var transaction IoTTransaction;
		err := json.Unmarshal(buf[0:n], &transaction)
		if err != nil {
			fmt.Println("Not a valid transaction...", err)
		}

		if transaction.From != nil {
			TxFee := [8]byte{}
			for index := range transaction.TxFee {
				TxFee[8-index-1] = byte(transaction.TxFee[index])
			}
			TxFee64, _ := binary.Uvarint(TxFee[:]);

			TxCnt := [4]byte{}
			for index := range transaction.TxCnt {
				TxCnt[4-index-1] = byte(transaction.TxCnt[index])
			}
			TxCnt32, _ := binary.Uvarint(TxCnt[:]);

			IotTx := protocol.IotTx{
				Header: byte(transaction.Header[0]),
				TxCnt:  uint32(TxCnt32),
				From:   protocol.SerializeHashContent(transaction.From),
				To:     protocol.SerializeHashContent(transaction.To),
				Sig:    transaction.Signature,
				Data:   transaction.Data,
				Fee:    TxFee64,
			}
			fmt.Println(IotTx)
			err = network.SendIotTx(util.Config.BootstrapIpport, &IotTx, p2p.IOTTX_BRDCST)
			if err != nil {
				logger.Printf("%v\n", err)
			}
		}
		logger.Println("Received transaction: ", string(buf[0:n]), " from ", addr)
	}
}
