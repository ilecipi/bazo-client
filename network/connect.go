package network

import (
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-client/util"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"log"
	"net"
	"strconv"
	"strings"
)

var (
	logger     *log.Logger
	peers      peersStruct
	register   = make(chan *peer)
	disconnect = make(chan *peer)
)

func Init() {
	logger = util.InitLogger()
	peers.minerConns = make(map[*peer]bool)

	go peerService()

	p, err := initiateNewClientConnection(storage.BOOTSTRAP_SERVER)
	if err != nil {
		logger.Fatal("Initiating new network connection failed: %v", err)
	}

	go minerConn(p)
}

func initiateNewClientConnection(dial string) (*peer, error) {
	var conn net.Conn

	//Open up a tcp dial and instantiate a peer struct, wait for adding it to the peerStruct before we finalize
	//the handshake
	conn, err := net.Dial("tcp", dial)
	if err != nil {
		return nil, err
	}

	p := newPeer(conn, strings.Split(dial, ":")[1])

	//Extracts the port from our localConn variable (which is in the form IP:Port)
	localPort, err := strconv.Atoi(strings.Split(util.LIGHT_CLIENT_SERVER_PORT, ":")[1])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Parsing port failed: %v\n", err))
	}

	packet, err := p2p.PrepareHandshake(p2p.CLIENT_PING, localPort)
	if err != nil {
		return nil, err
	}

	conn.Write(packet)

	//Wait for the other party to finish the handshake with the corresponding message
	header, _, err := rcvData(p)
	if err != nil || header.TypeID != p2p.CLIENT_PONG {
		return nil, errors.New(fmt.Sprintf("Failed to complete network handshake: %v", err))
	}

	return p, nil
}

func minerConn(p *peer) {
	logger.Printf("Adding a new miner: %v\n", p.getIPPort())

	//Give the peer a channel
	p.ch = make(chan []byte)

	//Register withe the broadcast service and start the additional writer
	register <- p

	for {
		header, payload, err := rcvData(p)
		if err != nil {
			logger.Printf("Miner disconnected: %v\n", err)

			//In case of a comm fail, disconnect cleanly from the broadcast service
			disconnect <- p
			return
		}

		processIncomingMsg(p, header, payload)
	}
}
