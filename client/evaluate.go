package client

import (
	"github.com/bazo-blockchain/bazo-client/network"
	"github.com/bazo-blockchain/bazo-miner/protocol"
)

func evaluateRelevantBlockHashes(pubKey [64]byte) (relevantBlockHashes [][32]byte) {
	pubKeyHash := protocol.SerializeHashContent(pubKey)
	for _, blockHeader := range blockHeaders {
		//Block is relevant if:
		//account is beneficary or
		//account is in bloomfilter (all addresses involved in acctx/fundstx) or
		//config state changed
		if blockHeader.NrConfigTx > 0 || (blockHeader.NrElementsBF > 0 && blockHeader.BloomFilter.Test(pubKeyHash[:])) {
			relevantBlockHashes = append(relevantBlockHashes, blockHeader.Hash)
		}
	}

	return relevantBlockHashes
}

func getRelevantBlocks(pubKey [64]byte) (relevantBlocks []*protocol.Block, err error) {
	for _, blockHash := range evaluateRelevantBlockHashes(pubKey) {
		err := network.BlockReq(blockHash[:])
		if err != nil {
			return nil, err
		}

		blockI, err := network.Fetch(network.BlockChan)
		if err != nil {
			return nil, err
		}

		var block *protocol.Block
		block = blockI.(*protocol.Block)
		relevantBlocks = append(relevantBlocks, block)
	}

	return relevantBlocks, nil
}
