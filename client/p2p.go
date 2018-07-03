package client

import (
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"net"
	"strings"
	"strconv"
)

const (
	LIGHT_CLIENT_SERVER_IP   = storage.BOOTSTRAP_SERVER_IP
	LIGHT_CLIENT_SERVER_PORT = ":8001"
	LIGHT_CLIENT_SERVER      = LIGHT_CLIENT_SERVER_IP + LIGHT_CLIENT_SERVER_PORT

	MULTISIG_SERVER_IP   = storage.BOOTSTRAP_SERVER_IP
	MULTISIG_SERVER_PORT = ":8002"
	MULTISIG_SERVER      = MULTISIG_SERVER_IP + MULTISIG_SERVER_PORT
)

func initiateNewClientConnection(dial string) (*p2p.Peer, error) {
	var conn net.Conn

	//Open up a tcp dial and instantiate a peer struct, wait for adding it to the peerStruct before we finalize
	//the handshake
	conn, err := net.Dial("tcp", dial)
	if err != nil {
		return nil, err
	}

	p := p2p.NewPeer(conn, strings.Split(dial, ":")[1], p2p.PEERTYPE_CLIENT)

	//Extracts the port from our localConn variable (which is in the form IP:Port)
	localPort, err := strconv.Atoi(strings.Split(LIGHT_CLIENT_SERVER_PORT, ":")[1])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Parsing port failed: %v\n", err))
	}

	packet, err := p2p.PrepareHandshake(p2p.CLIENT_PING, localPort)
	if err != nil {
		return nil, err
	}

	conn.Write(packet)

	//Wait for the other party to finish the handshake with the corresponding message
	header, _, err := p2p.RcvData(p)
	if err != nil || header.TypeID != p2p.CLIENT_PONG {
		return nil, errors.New(fmt.Sprintf("Failed to complete miner handshake: %v", err))
	}

	return p, nil
}

func reqBlock(blockHash [32]byte) (block *protocol.Block) {
	if conn := p2p.Connect(storage.BOOTSTRAP_SERVER); conn != nil {

		packet := p2p.BuildPacket(p2p.BLOCK_REQ, blockHash[:])
		conn.Write(packet)

		header, payload, err := p2p.RcvData_(conn)
		if err != nil || header.TypeID != p2p.BLOCK_RES {
			logger.Printf("Requesting block failed.")
			return
		}

		block = block.Decode(payload)

		conn.Close()
	}

	return block
}

func reqTx(txType uint8, txHash [32]byte) (tx protocol.Transaction) {
	if conn := p2p.Connect(storage.BOOTSTRAP_SERVER); conn != nil {

		packet := p2p.BuildPacket(txType, txHash[:])
		conn.Write(packet)

		header, payload, err := p2p.RcvData_(conn)
		if err != nil {
			logger.Printf("Requesting tx failed.")
			return
		}

		switch header.TypeID {
		case p2p.ACCTX_RES:
			var accTx *protocol.AccTx
			accTx = accTx.Decode(payload)
			tx = accTx
		case p2p.CONFIGTX_RES:
			var configTx *protocol.ConfigTx
			configTx = configTx.Decode(payload)
			tx = configTx
		case p2p.FUNDSTX_RES:
			var fundsTx *protocol.FundsTx
			fundsTx = fundsTx.Decode(payload)
			tx = fundsTx
		case p2p.STAKETX_RES:
			var stakeTx *protocol.StakeTx
			stakeTx = stakeTx.Decode(payload)
			tx = stakeTx
		}

		conn.Close()
	}

	return tx
}

func reqIntermediateNodes(blockHash [32]byte, txHash [32]byte) (nodes [][32]byte) {
	if conn := p2p.Connect(storage.BOOTSTRAP_SERVER); conn != nil {
		var data [][]byte
		data = append(data, blockHash[:])
		data = append(data, txHash[:])

		packet := p2p.BuildPacket(p2p.INTERMEDIATE_NODES_REQ, protocol.Encode(data, 32))
		conn.Write(packet)

		header, payload, err := p2p.RcvData_(conn)
		if err != nil || header.TypeID != p2p.INTERMEDIATE_NODES_RES {
			logger.Printf("Requesting intermediate nodes failed.")
			return
		}

		for _, data := range protocol.Decode(payload, 32) {
			var node [32]byte
			copy(node[:], data)
			nodes = append(nodes, node)
		}

		conn.Close()
	}

	return nil
}

func reqBlockHeader(blockHash []byte) (blockHeader *protocol.Block) {
	if conn := p2p.Connect(storage.BOOTSTRAP_SERVER); conn != nil {

		packet := p2p.BuildPacket(p2p.BLOCK_HEADER_REQ, blockHash)
		conn.Write(packet)

		header, payload, err := p2p.RcvData_(conn)
		if err != nil || header.TypeID != p2p.BlOCK_HEADER_RES {
			logger.Printf("Requesting block header failed.")
			return
		}

		blockHeader = blockHeader.DecodeHeader(payload)

		conn.Close()
	}

	return blockHeader
}

func ReqAcc(accountHash [32]byte) (acc *protocol.Account) {
	if conn := p2p.Connect(storage.BOOTSTRAP_SERVER); conn != nil {

		packet := p2p.BuildPacket(p2p.ACC_REQ, accountHash[:])
		conn.Write(packet)

		header, payload, err := p2p.RcvData_(conn)
		if err != nil || header.TypeID != p2p.ACC_RES {
			logger.Printf("Requesting account failed.")
			return nil
		}

		acc = acc.Decode(payload)

		conn.Close()
	}

	return acc
}

func reqRootAcc(accountHash [32]byte) (rootAcc *protocol.Account) {
	if conn := p2p.Connect(storage.BOOTSTRAP_SERVER); conn != nil {

		packet := p2p.BuildPacket(p2p.ROOTACC_REQ, accountHash[:])
		conn.Write(packet)

		_, payload, err := p2p.RcvData_(conn)
		if err != nil {
			logger.Printf("Requesting root account failed.")
			return nil
		}

		rootAcc = rootAcc.Decode(payload)

		conn.Close()
	}

	return rootAcc
}

func reqNonVerifiedTx(addressHash [32]byte) (nonVerifiedTxs []*protocol.FundsTx) {
	if conn := p2p.Connect(MULTISIG_SERVER); conn != nil {
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

func SendTx(dial string, tx protocol.Transaction, typeID uint8) (err error) {
	if conn := p2p.Connect(dial); conn != nil {
		packet := p2p.BuildPacket(typeID, tx.Encode())
		conn.Write(packet)

		header, payload, err := p2p.RcvData_(conn)
		if err != nil || header.TypeID == p2p.NOT_FOUND {
			err = errors.New(string(payload[:]))
		}
		conn.Close()

		return err
	}

	txHash := tx.Hash()
	return errors.New(fmt.Sprintf("Sending tx %x failed.", txHash[:8]))
}
