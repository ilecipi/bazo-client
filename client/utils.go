package client

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"log"
	"net"
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

func put(slice []*FundsTxJson, tx *FundsTxJson) {
	for i := 0; i < 9; i++ {
		slice[i] = slice[i+1]
	}

	slice[9] = tx
}
