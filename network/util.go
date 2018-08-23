package network

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-client/util"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"time"
)

func Fetch(channelToFetchFrom chan interface{}) (payload interface{}, err error) {
	select {
	case payload = <-channelToFetchFrom:
	case <-time.After(util.FETCH_TIMEOUT * time.Second):
		return nil, errors.New("Fetching timed out.")
	}

	return payload, nil
}

func Fetch32Bytes(channelToFetchFrom chan [][32]byte) (payload [][32]byte, err error) {
	select {
	case payload = <-channelToFetchFrom:
	case <-time.After(util.FETCH_TIMEOUT * time.Second):
		return nil, errors.New("Fetching timed out.")
	}

	return payload, nil
}

func rcvData(p *peer) (header *p2p.Header, payload []byte, err error) {
	reader := bufio.NewReader(p.conn)
	header, err = readHeader(reader)

	if err != nil {
		p.conn.Close()
		return nil, nil, errors.New(fmt.Sprintf("Connection to %v aborted: %v", p.getIPPort(), err))
	}
	payload = make([]byte, header.Len)

	for cnt := 0; cnt < int(header.Len); cnt++ {
		payload[cnt], err = reader.ReadByte()
		if err != nil {
			p.conn.Close()
			return nil, nil, errors.New(fmt.Sprintf("Connection to %v aborted: %v", p.getIPPort(), err))
		}
	}

	logger.Printf("Receive message:\nSender: %v\nType: %v\nPayload length: %v\n", p.getIPPort(), p2p.LogMapping[header.TypeID], len(payload))

	return header, payload, nil
}

func readHeader(reader *bufio.Reader) (*p2p.Header, error) {
	//the first four bytes of any incoming messages is the length of the payload
	//error catching after every read is necessary to avoid panicking
	var headerArr [p2p.HEADER_LEN]byte
	//reading byte by byte is surprisingly fast and works a lot better for concurrent connections
	for i := range headerArr {
		extr, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		headerArr[i] = extr
	}

	header := extractHeader(headerArr[:])
	return header, nil
}

//Decoupled functionality for testing reasons
func extractHeader(headerData []byte) *p2p.Header {

	header := new(p2p.Header)

	lenBuf := [4]byte{headerData[0], headerData[1], headerData[2], headerData[3]}
	packetLen := binary.BigEndian.Uint32(lenBuf[:])

	header.Len = packetLen
	header.TypeID = uint8(headerData[4])
	return header
}

func sendData(p *peer, payload []byte) {
	logger.Printf("Send message:\nReceiver: %v\nType: %v\nPayload length: %v\n", p.getIPPort(), p2p.LogMapping[payload[4]], len(payload)-p2p.HEADER_LEN)

	p.l.Lock()
	p.conn.Write(payload)
	p.l.Unlock()
}
