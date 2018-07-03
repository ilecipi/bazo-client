package client

import (
	"github.com/bazo-blockchain/bazo-miner/miner"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"time"
	"github.com/bazo-blockchain/bazo-miner/storage"
)

var (
	//All blockheaders of the whole chain
	blockHeaders     []*protocol.Block
	cnt              uint32
	activeParameters miner.Parameters
	UnsignedAccTx    = make(map[[32]byte]*protocol.AccTx)
	UnsignedConfigTx = make(map[[32]byte]*protocol.ConfigTx)
	UnsignedFundsTx  = make(map[[32]byte]*protocol.FundsTx)
)

//Load initially all block headers and invert them (first oldest, last latest)
func InitState() {
	_, err := initiateNewClientConnection(storage.BOOTSTRAP_SERVER)
	if err != nil {
		logger.Fatal("Initiating new miner connection failed: %v", err)
	}

	cnt = 0
	updateBlockHeader(true)

	go refreshState()
}

//Update allBlockHeaders to the latest header
func refreshState() {
	for {
		time.Sleep(10 * time.Second)

		updateBlockHeader(false)
	}
}

func updateBlockHeader(initial bool) {
	var loaded []*protocol.Block
	if youngest := reqBlockHeader(nil); youngest == nil {
		logger.Printf("Refreshing state failed.")
	} else {
		if len(blockHeaders) > 0 {
			loaded = checkForNewBlockHeaders(initial, youngest, blockHeaders[len(blockHeaders)-1].Hash, loaded)
		} else {
			loaded = checkForNewBlockHeaders(initial, youngest, [32]byte{}, loaded)
		}

		blockHeaders = append(blockHeaders, loaded...)
	}
}

//Get new blockheaders recursively
func checkForNewBlockHeaders(initial bool, latest *protocol.Block, lastLoaded [32]byte, loaded []*protocol.Block) []*protocol.Block {
	if latest.Hash != lastLoaded {

		if initial {
			logger.Printf("Header %v loaded\n", cnt)
			cnt++
		} else {
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

		var ancestor *protocol.Block
		if ancestor = reqBlockHeader(latest.PrevHash[:]); ancestor == nil {
			logger.Printf("Refreshing state failed.")
		}

		loaded = checkForNewBlockHeaders(initial, ancestor, lastLoaded, loaded)
		loaded = append(loaded, latest)
	}

	return loaded
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

	addressHash := protocol.SerializeHashContent(acc.Address)
	for _, tx := range reqNonVerifiedTx(addressHash) {
		if tx.To == addressHash {
			put(lastTenTx, ConvertFundsTx(tx, "not verified"))
		}
		if tx.From == addressHash {
			acc.TxCnt++
		}
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
