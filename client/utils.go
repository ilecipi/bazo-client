package client

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"golang.org/x/crypto/sha3"
	"log"
	"math/big"
	"net"
	"os"
	"strings"
	"time"
)

func Connect(connectionString string) *net.TCPConn {
	tcpAddr, err := net.ResolveTCPAddr("tcp", connectionString)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatal(err)
	}

	conn.SetLinger(0)
	conn.SetDeadline(time.Now().Add(20 * time.Second))

	return conn
}

func SerializeHashContent(data interface{}) (hash [32]byte) {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, data)
	return sha3.Sum256(buf.Bytes())
}

func getKeys(keyFile string) (myPubKey [64]byte, myPubKeyHash [32]byte, err error) {
	myKeys, err := os.Open(keyFile)
	if err != nil {
		return myPubKey, myPubKeyHash, err
	}

	reader := bufio.NewReader(myKeys)

	//We only need the public key
	pub1, _ := reader.ReadString('\n')
	pub2, _ := reader.ReadString('\n')

	pub1Int, _ := new(big.Int).SetString(strings.Split(pub1, "\n")[0], 16)
	pub2Int, _ := new(big.Int).SetString(strings.Split(pub2, "\n")[0], 16)

	copy(myPubKey[0:32], pub1Int.Bytes())
	copy(myPubKey[32:64], pub2Int.Bytes())

	myPubKeyHash = SerializeHashContent(myPubKey)

	return myPubKey, myPubKeyHash, err
}

func RcvData(c net.Conn) (header *p2p.Header, payload []byte, err error) {
	reader := bufio.NewReader(c)
	header, err = p2p.ReadHeader(reader)
	if err != nil {
		c.Close()
		return nil, nil, errors.New(fmt.Sprintf("Connection to aborted: (%v)\n", err))
	}
	payload = make([]byte, header.Len)

	for cnt := 0; cnt < int(header.Len); cnt++ {
		payload[cnt], err = reader.ReadByte()
		if err != nil {
			c.Close()
			return nil, nil, errors.New(fmt.Sprintf("Connection to aborted: %v\n", err))
		}
	}

	return header, payload, nil
}
