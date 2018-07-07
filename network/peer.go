package network

import (
	"math/rand"
	"net"
	"strings"
	"sync"
)

//The reason we use an additional listener port is because the port the miner connected to this peer
//is not the same as the one it listens to for new connections. When we are queried for neighbors
//we send the IP address in p.conn.RemotAddr() with the listenerPort.
type peer struct {
	conn         net.Conn
	ch           chan []byte
	l            sync.Mutex
	listenerPort string
}

//Block constructor, argument is the previous block in the blockchain.
func newPeer(conn net.Conn, listenerPort string) *peer {
	p := new(peer)
	p.conn = conn
	p.ch = nil
	p.l = sync.Mutex{}
	p.listenerPort = listenerPort

	return p
}

//PeerStruct is a thread-safe map that supports all necessary map operations needed by the server.
type peersStruct struct {
	minerConns map[*peer]bool
	peerMutex  sync.Mutex
}

func (p *peer) getIPPort() string {
	ip := strings.Split(p.conn.RemoteAddr().String(), ":")
	//Cut off original port.
	port := p.listenerPort

	return ip[0] + ":" + port
}

func (peers peersStruct) add(p *peer) {
	peers.peerMutex.Lock()
	defer peers.peerMutex.Unlock()

	peers.minerConns[p] = true
}

func (peers peersStruct) delete(p *peer) {
	peers.peerMutex.Lock()
	defer peers.peerMutex.Unlock()

	delete(peers.minerConns, p)
}

func (peers peersStruct) len(peerType uint) (length int) {
	length = len(peers.minerConns)

	return length
}

func (peers peersStruct) getRandomPeer() (p *peer) {
	//Acquire list before locking, otherwise deadlock
	peerList := peers.getAllPeers()

	if len(peerList) == 0 {
		return nil
	} else {
		return peerList[int(rand.Uint32())%len(peerList)]
	}
}

func (peers peersStruct) getAllPeers() []*peer {
	peers.peerMutex.Lock()
	defer peers.peerMutex.Unlock()

	var peerList []*peer

	for p := range peers.minerConns {
		peerList = append(peerList, p)
	}

	return peerList
}
