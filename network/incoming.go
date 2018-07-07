package network

import (
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
)

var (
	BlockHeaderIn = make(chan *protocol.Block)

	BlockChan       = make(chan interface{})
	BlockHeaderChan = make(chan interface{})
	FundsTxChan     = make(chan interface{})
	AccTxChan       = make(chan interface{})
	ConfigTxChan    = make(chan interface{})
	StakeTxChan     = make(chan interface{})
)

func processIncomingMsg(p *peer, header *p2p.Header, payload []byte) {
	switch header.TypeID {
	case p2p.BLOCK_HEADER_BRDCST:
		blockHeaderBrdcst(p, payload)
	case p2p.BLOCK_RES:
		blockRes(p, payload)
	case p2p.BlOCK_HEADER_RES:
		blockHeaderRes(p, payload)
	case p2p.FUNDSTX_RES:
		txRes(p, payload, p2p.FUNDSTX_RES)
	case p2p.ACCTX_RES:
		txRes(p, payload, p2p.ACCTX_RES)
	case p2p.CONFIGTX_RES:
		txRes(p, payload, p2p.CONFIGTX_RES)
	case p2p.STAKETX_RES:
		txRes(p, payload, p2p.STAKETX_RES)
	}
}
