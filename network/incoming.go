package network

import (
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
)

var (
	Uptodate      = false
	BlockHeaderIn = make(chan *protocol.Block)
	iplistChan    = make(chan string, p2p.MIN_MINERS)

	BlockChan             = make(chan interface{})
	BlockHeaderChan       = make(chan interface{})
	FundsTxChan           = make(chan interface{})
	AccTxChan             = make(chan interface{})
	ConfigTxChan          = make(chan interface{})
	StakeTxChan           = make(chan interface{})
	AccChan               = make(chan interface{})
	IntermediateNodesChan = make(chan [][32]byte)
)

func processIncomingMsg(p *peer, header *p2p.Header, payload []byte) {
	switch header.TypeID {
	//BROADCAST
	case p2p.BLOCK_HEADER_BRDCST:
		if Uptodate {
			blockHeaderBrdcst(p, payload)
		} else {
			logger.Println("Broadcastet block header not processed.")
		}

		//RESULTS
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
	case p2p.ACC_RES:
		accRes(p, payload)
	case p2p.ROOTACC_RES:
		accRes(p, payload)
	case p2p.INTERMEDIATE_NODES_RES:
		intermediateNodesRes(p, payload)
	case p2p.NEIGHBOR_RES:
		processNeighborRes(p, payload)
	}
}
