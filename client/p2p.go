package client

import (
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
)

func reqNonVerifiedTx(addressHash [32]byte) (nonVerifiedTxs []*protocol.FundsTx) {
	//TODO Revise connection to Multisig server
	if conn := p2p.Connect(""); conn != nil {
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
