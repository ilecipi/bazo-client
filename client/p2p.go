package client

import (
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
)

func reqBlock(blockHash [32]byte) (block *protocol.Block) {

	conn := Connect(p2p.BOOTSTRAP_SERVER)

	packet := p2p.BuildPacket(p2p.BLOCK_REQ, blockHash[:])
	conn.Write(packet)

	header, payload, err := rcvData(conn)
	if err != nil {
		logger.Printf("Disconnected: %v\n", err)
		return
	}

	if header.TypeID == p2p.BLOCK_RES {
		block = block.Decode(payload)
	}

	conn.Close()

	return block
}

func reqTx(txType uint8, txHash [32]byte) (tx protocol.Transaction) {

	conn := Connect(p2p.BOOTSTRAP_SERVER)

	packet := p2p.BuildPacket(txType, txHash[:])
	conn.Write(packet)

	header, payload, err := rcvData(conn)
	if err != nil {
		logger.Printf("Disconnected: %v\n", err)
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

	return tx
}

func reqIntermediateNodes(blockHash [32]byte, txHash [32]byte) (nodes [][32]byte) {
	conn := Connect(p2p.BOOTSTRAP_SERVER)

	packet := p2p.BuildPacket(p2p.INTERMEDIATE_NODES_REQ, protocol.SerializeSlice32([][32]byte{blockHash, txHash}))
	conn.Write(packet)

	header, payload, err := rcvData(conn)
	if err != nil {
		logger.Printf("Disconnected: %v\n", err)
		return
	}

	if header.TypeID == p2p.INTERMEDIATE_NODES_RES {
		return protocol.DeserializeSlice32(payload)
	}

	conn.Close()

	return nil
}

//Request blockheader from network
func reqBlockHeader(blockHash []byte) (blockHeader *protocol.Block) {
	conn := Connect(p2p.BOOTSTRAP_SERVER)

	packet := p2p.BuildPacket(p2p.BLOCK_HEADER_REQ, blockHash)
	conn.Write(packet)

	header, payload, err := rcvData(conn)
	if err != nil {
		logger.Printf("Disconnected: %v\n", err)
		return
	}

	if header.TypeID == p2p.BlOCK_HEADER_RES {
		blockHeader = blockHeader.Decode(payload)
	}

	conn.Close()

	return blockHeader
}

//Check if our address is the initial root account, since for it no accTx exists
func reqRootAcc(accountHash [32]byte) (rootAcc *protocol.Account) {
	conn := Connect(p2p.BOOTSTRAP_SERVER)

	packet := p2p.BuildPacket(p2p.ROOTACC_REQ, accountHash[:])
	conn.Write(packet)

	header, payload, err := rcvData(conn)
	if err != nil {
		logger.Printf("Disconnected: %v\n", err)
		return nil
	}

	if header.TypeID == p2p.ROOTACC_RES {
		rootAcc = rootAcc.Decode(payload)
	}

	conn.Close()

	return rootAcc
}

func SendTx(tx protocol.Transaction, typeID uint8) (err error) {
	//Transaction creation successful
	packet := p2p.BuildPacket(typeID, tx.Encode())

	//Open a connection
	conn := Connect(p2p.BOOTSTRAP_SERVER)
	conn.Write(packet)

	header, _, err := rcvData(conn)
	if header.TypeID != p2p.TX_BRDCST_ACK || err != nil {
		err = errors.New(fmt.Sprintf("%v\nCould not send the following transaction: %x", err, tx.Hash()))
	}

	conn.Close()

	return err
}
