package client

import (
	"github.com/bazo-blockchain/bazo-miner/miner"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"time"
)

var (
	//All blockheaders of the whole chain
	allBlockHeaders  []*protocol.Block
	activeParameters miner.Parameters
	UnsignedAccTx    = make(map[[32]byte]*protocol.AccTx)
	UnsignedConfigTx = make(map[[32]byte]*protocol.ConfigTx)
	UnsignedFundsTx  = make(map[[32]byte]*protocol.FundsTx)
)

//Load initially all block headers and invert them (first oldest, last latest)
func InitState() {
	loadAllBlockHeaders()
	allBlockHeaders = miner.InvertBlockArray(allBlockHeaders)

	go refreshState()
}

//Update allBlockHeaders to the latest header
func refreshState() {
	for {
		//Try to load all headers if non have been loaded before
		if len(allBlockHeaders) > 0 {
			var newBlockHeaders []*protocol.Block
			if lastBlockHeader := reqBlockHeader(nil); lastBlockHeader == nil {
				logger.Printf("Refreshing state failed.")
			} else {
				newBlockHeaders = getNewBlockHeaders(lastBlockHeader, allBlockHeaders[len(allBlockHeaders)-1], newBlockHeaders)
				allBlockHeaders = append(allBlockHeaders, newBlockHeaders...)
			}
		} else {
			loadAllBlockHeaders()
		}

		time.Sleep(10 * time.Second)
	}
}

//Get new blockheaders recursively
func getNewBlockHeaders(latest *protocol.Block, eldest *protocol.Block, list []*protocol.Block) []*protocol.Block {
	if latest.Hash != eldest.Hash {
		ancestor := reqBlockHeader(latest.PrevHash[:])
		list = getNewBlockHeaders(ancestor, eldest, list)
		list = append(list, latest)
		logger.Printf("Header: %x loaded\n"+
			"NrFundsTx: %v\n"+
			"NrAccTx: %v\n"+
			"NrConfigTx: %v\n"+
			"NrStakeTx: %v\n",
			latest.Hash[:8],
			latest.NrFundsTx,
			latest.NrAccTx,
			latest.NrConfigTx,
			latest.NrConfigTx)
	}

	return list
}

func getState(acc *Account, lastTenTx []*FundsTxJson) error {
	pubKeyHash := protocol.SerializeHashContent(acc.Address)

	//Get blocks if the Acc address:
	//* got issued as an Acc
	//* sent funds
	//* received funds
	//* is block's beneficiary
	//* nr of configTx in block is > 0 (in order to maintain params in light-client)
	relevantBlocks := getRelevantBlocks(acc.Address)

	for _, block := range relevantBlocks {
		if block != nil {
			//Collect block reward
			if block.Beneficiary == pubKeyHash {
				acc.Balance += activeParameters.Block_reward
			}

			//Balance funds and collect fee
			for _, txHash := range block.FundsTxData {
				tx := reqTx(p2p.FUNDSTX_REQ, txHash)
				fundsTx := tx.(*protocol.FundsTx)

				if fundsTx.From == pubKeyHash || fundsTx.To == pubKeyHash || block.Beneficiary == pubKeyHash {
					//Validate tx
					if err := validateTx(block, tx, txHash); err != nil {
						return err
					}

					if fundsTx.From == pubKeyHash {
						//If Acc is no root, balance funds
						if !acc.IsRoot {
							acc.Balance -= fundsTx.Amount
							acc.Balance -= fundsTx.Fee
						}

						acc.TxCnt += 1
					}

					if fundsTx.To == pubKeyHash {
						acc.Balance += fundsTx.Amount

						put(lastTenTx, ConvertFundsTx(fundsTx, "verified"))
					}

					if block.Beneficiary == pubKeyHash {
						acc.Balance += fundsTx.Fee
					}
				}
			}

			//Check if Account was issued and collect fee
			for _, txHash := range block.AccTxData {
				tx := reqTx(p2p.ACCTX_REQ, txHash)
				accTx := tx.(*protocol.AccTx)

				if accTx.PubKey == acc.Address || block.Beneficiary == pubKeyHash {
					//Validate tx
					if err := validateTx(block, tx, txHash); err != nil {
						return err
					}

					if accTx.PubKey == acc.Address {
						acc.IsCreated = true
					}

					if block.Beneficiary == pubKeyHash {
						acc.Balance += accTx.Fee
					}
				}
			}

			//Update config parameters and collect fee
			for _, txHash := range block.ConfigTxData {
				tx := reqTx(p2p.CONFIGTX_REQ, txHash)
				configTx := tx.(*protocol.ConfigTx)
				configTxSlice := []*protocol.ConfigTx{configTx}

				if block.Beneficiary == pubKeyHash {
					//Validate tx
					if err := validateTx(block, tx, txHash); err != nil {
						return err
					}

					acc.Balance += configTx.Fee
				}

				miner.CheckAndChangeParameters(&activeParameters, &configTxSlice)
			}

			//TODO stakeTx

		}
	}

	for _, tx := range reqNonVerifiedTx(protocol.SerializeHashContent(acc.Address)) {
		put(lastTenTx, ConvertFundsTx(tx, "not verified"))
	}

	return nil
}

func getRelevantBlocks(pubKey [64]byte) (relevantBlocks []*protocol.Block) {
	for _, blockHash := range getRelevantBlockHashes(pubKey) {
		if block := reqBlock(blockHash); block != nil {
			relevantBlocks = append(relevantBlocks, block)
		}
	}

	return relevantBlocks
}

func getRelevantBlockHashes(pubKey [64]byte) (relevantBlockHashes [][32]byte) {
	pubKeyHash := protocol.SerializeHashContent(pubKey)
	for _, blockHeader := range allBlockHeaders {
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

//Returns all block headers, youngest first, genesis last
func loadAllBlockHeaders() {
	logger.Println("Loading all block headers.")

	counter := 0

	//If no blockhash as param is given, the last block header is given back
	if blockHeader := reqBlockHeader(nil); blockHeader != nil {
		allBlockHeaders = append(allBlockHeaders, blockHeader)
		logger.Printf("Header %v loaded", counter)
		counter++
		prevHash := blockHeader.PrevHash

		for blockHeader.Hash != [32]byte{} {
			if blockHeader = reqBlockHeader(prevHash[:]); blockHeader != nil {
				allBlockHeaders = append(allBlockHeaders, blockHeader)
				logger.Printf("Header %v loaded", counter)
				counter++
				prevHash = blockHeader.PrevHash
			} else {
				logger.Printf("Loading block headers failed.")
			}
		}

		logger.Println("All block headers loaded.")
	} else {
		logger.Printf("Loading block headers failed.")
	}
}
