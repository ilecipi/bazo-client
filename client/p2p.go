package client

import (
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-client/util"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
)

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
	if conn := p2p.Connect(util.MULTISIG_SERVER); conn != nil {
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
