package network

import (
	"errors"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-client/util"
)

func BlockReq(blockHash []byte) error {
	p := peers.getRandomPeer()
	if p == nil {
		return errors.New("Couldn't get a connection, request not transmitted.")
	}

	packet := p2p.BuildPacket(p2p.BLOCK_REQ, blockHash[:])
	sendData(p, packet)

	return nil
}

func BlockHeaderReq(blockHash []byte) error {
	p := peers.getRandomPeer()
	if p == nil {
		return errors.New("Couldn't get a connection, request not transmitted.")
	}

	packet := p2p.BuildPacket(p2p.BLOCK_HEADER_REQ, blockHash[:])
	sendData(p, packet)

	return nil
}

func TxReq(txType uint8, txHash [32]byte) error {
	p := peers.getRandomPeer()
	if p == nil {
		return errors.New("Couldn't get a connection, request not transmitted.")
	}

	packet := p2p.BuildPacket(txType, txHash[:])
	sendData(p, packet)

	return nil
}

func AccReq(root bool, addressHash [32]byte) error {
	p := peers.getRandomPeer()
	if p == nil {
		return errors.New("Couldn't get a connection, request not transmitted.")
	}

	var packet []byte
	if root {
		packet = p2p.BuildPacket(p2p.ROOTACC_REQ, addressHash[:])
	} else {
		packet = p2p.BuildPacket(p2p.ACC_REQ, addressHash[:])
	}

	sendData(p, packet)

	return nil
}

func NonVerifiedTxReq(addressHash [32]byte) (nonVerifiedTxs []*protocol.FundsTx) {
	if conn := p2p.Connect(util.Config.MultisigIpport); conn != nil {
		packet := p2p.BuildPacket(p2p.FUNDSTX_REQ, addressHash[:])
		conn.Write(packet)

		header, payload, err := p2p.RcvData_(conn)
		if err != nil || header.TypeID != p2p.FUNDSTX_RES {
			logger.Printf("Requesting non verified tx failed.")
			return nil
		}

		for _, data := range protocol.Decode(payload, protocol.FUNDSTX_SIZE) {
			var tx *protocol.FundsTx
			nonVerifiedTxs = append(nonVerifiedTxs, tx.Decode(data))
		}
	}

	return nonVerifiedTxs
}

func IntermediateNodesReq(blockHash [32]byte, txHash [32]byte) error {
	p := peers.getRandomPeer()
	if p == nil {
		return errors.New("Couldn't get a connection, request not transmitted.")
	}

	var data [][]byte
	data = append(data, blockHash[:])
	data = append(data, txHash[:])

	packet := p2p.BuildPacket(p2p.INTERMEDIATE_NODES_REQ, protocol.Encode(data, 32))
	sendData(p, packet)

	return nil
}

func neighborReq() {
	p := peers.getRandomPeer()
	if p == nil {
		logger.Print("Could not fetch a random peer.\n")
		return
	}

	packet := p2p.BuildPacket(p2p.NEIGHBOR_REQ, nil)
	sendData(p, packet)
}