package UDP

import (
	"encoding/binary"
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

func Init() {
	logger = util.InitLogger()
	logger.Printf("%v\n\n", "Listening to incoming UDP packets...")
	ServerConn, _ := net.ListenUDP("udp", &net.UDPAddr{IP: []byte{0, 0, 0, 0}, Port: 5001, Zone: ""})

	defer ServerConn.Close()
	buf := make([]byte, 1024*2)
	for {
		n, addr, _ := ServerConn.ReadFromUDP(buf)

		logger.Println("Received transaction: from ", n)
		length, tx :=convert_tx(buf,n )

		var signature [64]byte;
		copy(signature[:],tx[:64]);
		tx = tx[64:]

		var To [32]byte;
		copy(To[:],tx[:32]);
		tx = tx[32:]

		var From [32]byte;
		copy(From[:],tx[:32]);
		tx = tx[32:]

		TxCnt := [4]byte{}
		for index := range tx[:4] {
			TxCnt[4-index-1] = byte(tx[index])
		}
		TxCnt32, _ := binary.Uvarint(TxCnt[:]);
		tx = tx[4:];

		TxFee := [8]byte{}
		for index := range tx[:8] {
			TxFee[8-index-1] = byte(tx[index])
		}
		TxFee64, _ := binary.Uvarint(TxFee[:]);
		tx = tx[8:];

		Header := tx[:1];
		tx = tx[1:];

		data := tx[:];

		IotTx := protocol.IotTx{
			Header: Header[0],
			TxCnt:  uint32(TxCnt32),
			From:   protocol.SerializeHashContent(From),
			To:     protocol.SerializeHashContent(To),
			Sig:    signature,
			Data:   data,
			Fee:    TxFee64,
		}
		err := network.SendIotTx(util.Config.BootstrapIpport, &IotTx, p2p.IOTTX_BRDCST)
		if err != nil && addr !=nil  {
			//logger.Printf("%v\n", err, addr)
		}
		fmt.Println("Length Overhead ", n," - ", length)
	}
}

func convert_tx(str []byte, length int) (int, []byte){
	tx_raw := [2000]byte{};
	i:=0;
	j:=0;

	for i < length {
		if(str[i] > 2) {
			tx_raw[j] = str[i];
			i++;
			j++;
		}
		if(str[i] == 0x1) {
			if(i+1 < length) {
				if(str[i+1] == 0x1) {
					tx_raw[j] = 0x0;
					i += 2;
					j++;
				} else if(str[i+1] == 0x2) {
					tx_raw[j] = 0x1;
					i += 2;
					j++;
				} else {
					// error
					i++;
					j++;

				}
			}
		}
		if(str[i] == 0x2) {
			if(i+1 < length) {
				if(str[i+1] == 0x2) {
					tx_raw[j] = 0x2;
					i += 2;
					j++;
				} else {
					// error
					i++;
					j++;
				}
			}
		}
	}
	return j,tx_raw[:j]

}
