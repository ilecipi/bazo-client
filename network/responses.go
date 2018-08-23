package network

import (
	"encoding/binary"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"strconv"
)

func blockHeaderBrdcst(p *peer, payload []byte) {
	var blockHeader *protocol.Block
	blockHeader = blockHeader.DecodeHeader(payload)
	BlockHeaderIn <- blockHeader
}

func blockRes(p *peer, payload []byte) {
	var blockHeader *protocol.Block
	blockHeader = blockHeader.Decode(payload)
	BlockChan <- blockHeader
}

func blockHeaderRes(p *peer, payload []byte) {
	var blockHeader *protocol.Block
	blockHeader = blockHeader.DecodeHeader(payload)
	BlockHeaderChan <- blockHeader
}

func txRes(p *peer, payload []byte, txType uint8) {
	if payload == nil {
		return
	}

	switch txType {
	case p2p.FUNDSTX_RES:
		var fundsTx *protocol.FundsTx
		fundsTx = fundsTx.Decode(payload)
		if fundsTx == nil {
			return
		}
		FundsTxChan <- fundsTx
	case p2p.ACCTX_RES:
		var accTx *protocol.AccTx
		accTx = accTx.Decode(payload)
		if accTx == nil {
			return
		}
		AccTxChan <- accTx
	case p2p.CONFIGTX_RES:
		var configTx *protocol.ConfigTx
		configTx = configTx.Decode(payload)
		if configTx == nil {
			return
		}
		ConfigTxChan <- configTx
	case p2p.STAKETX_RES:
		var stakeTx *protocol.StakeTx
		stakeTx = stakeTx.Decode(payload)
		if stakeTx == nil {
			return
		}
		StakeTxChan <- stakeTx
	}
}

func accRes(p *peer, payload []byte) {
	var acc *protocol.Account
	acc = acc.Decode(payload)

	AccChan <- acc
}

func intermediateNodesRes(p *peer, payload []byte) {
	var nodes [][32]byte
	for _, data := range protocol.Decode(payload, 32) {
		var node [32]byte
		copy(node[:], data)
		nodes = append(nodes, node)
	}

	IntermediateNodesChan <- nodes
}

func processNeighborRes(p *peer, payload []byte) {

	//Parse the incoming ipv4 addresses.
	ipportList := _processNeighborRes(payload)

	for _, ipportIter := range ipportList {
		logger.Printf("IP/Port received: %v\n", ipportIter)
		//iplistChan is a buffered channel to handle ips asynchronously.
		iplistChan <- ipportIter
	}
}

//Split the processNeighborRes function in two for cleaner testing.
func _processNeighborRes(payload []byte) (ipportList []string) {
	index := 0
	for cnt := 0; cnt < len(payload)/(p2p.IPV4ADDR_SIZE+p2p.PORT_SIZE); cnt++ {
		var addr string
		for singleAddr := index; singleAddr < index+p2p.IPV4ADDR_SIZE; singleAddr++ {
			tmp := int(payload[singleAddr])
			addr += strconv.Itoa(tmp)
			addr += "."
		}
		//Remove trailing dot.
		addr = addr[:len(addr)-1]
		addr += ":"
		//Extract port number.
		addr += strconv.Itoa(int(binary.BigEndian.Uint16(payload[index+4 : index+6])))

		ipportList = append(ipportList, addr)
		index += p2p.IPV4ADDR_SIZE + p2p.PORT_SIZE
	}
	return ipportList
}
