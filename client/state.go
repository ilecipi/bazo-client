package client

import (
	"fmt"
	"github.com/bazo-blockchain/bazo-client/network"
	"github.com/bazo-blockchain/bazo-miner/miner"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"time"
	"github.com/bazo-blockchain/bazo-client/cstorage"
)

var (
	//All blockheaders of the whole chain
	blockHeaders []*protocol.Block

	activeParameters miner.Parameters

	UnsignedAccTx    = make(map[[32]byte]*protocol.AccTx)
	UnsignedConfigTx = make(map[[32]byte]*protocol.ConfigTx)
	UnsignedFundsTx  = make(map[[32]byte]*protocol.FundsTx)
)

//Update allBlockHeaders to the latest header. Start listening to broadcasted headers after.
func sync() {
	updateBlockHeaders()

	go incomingBlockHeaders()
}

func updateBlockHeaders() {
	//Wait until a connection is to the network is established.
	time.Sleep(10 * time.Second)

	var loaded []*protocol.Block
	var youngest *protocol.Block

	youngest = loadBlockHeader(nil)
	if youngest == nil {
		logger.Fatal()
	} else {
		loaded = checkForNewBlockHeaders(youngest, [32]byte{}, loaded)
	}

	blockHeaders = append(blockHeaders, loaded...)

	//The client is up to date with the network and can start listening for incoming headers.
	network.Uptodate = true
}

//Load all blockheaders from latest to the lastloaded (hash) given recursively.
func checkForNewBlockHeaders(latest *protocol.Block, lastLoaded [32]byte, loaded []*protocol.Block) []*protocol.Block {
	if latest.Hash != lastLoaded {
		var ancestor *protocol.Block

		if ancestor = cstorage.ReadBlockHeader(latest.PrevHash); ancestor == nil {
			ancestor = loadBlockHeader(latest.PrevHash[:])
		}

		if ancestor == nil {
			//Try again
			ancestor = latest
		} else {
			cstorage.WriteBlockHeader(ancestor)

			logger.Printf("Header %x with height %v loaded\n",
				ancestor.Hash[:8],
				ancestor.Height)
		}

		loaded = checkForNewBlockHeaders(ancestor, lastLoaded, loaded)
		loaded = append(loaded, latest)
	}

	return loaded
}

func loadBlockHeader(blockHash []byte) (blockHeader *protocol.Block) {
	var errormsg string
	if blockHash != nil {
		errormsg = fmt.Sprintf("Loading block header %x failed: ", blockHash[:8])
	}

	err := network.BlockHeaderReq(blockHash[:])
	if err != nil {
		logger.Println(errormsg + err.Error())
		return nil
	}

	blockHeaderI, err := network.Fetch(network.BlockHeaderChan)
	if err != nil {
		logger.Println(errormsg + err.Error())
		return nil
	}

	blockHeader = blockHeaderI.(*protocol.Block)

	return blockHeader
}

func incomingBlockHeaders() {
	for {
		blockHeaderIn := <-network.BlockHeaderIn

		cstorage.WriteBlockHeader(blockHeaderIn)
		cstorage.WriteLastBlockHeader(blockHeaderIn)

		//The incoming block header is already the last saved in the array.
		if blockHeaderIn.Hash == blockHeaders[len(blockHeaders)-1].Hash {
			break
		}

		if blockHeaderIn.PrevHash == blockHeaders[len(blockHeaders)-1].Hash {
			blockHeaders = append(blockHeaders, blockHeaderIn)
		} else {
			//The client is out of sync. Header cannot be appended to the array. The client must sync first.
			//Set the uptodate flag to false in order to avoid listening to new incoming block headers.
			network.Uptodate = false
			blockHeaders = checkForNewBlockHeaders(blockHeaderIn, blockHeaders[len(blockHeaders)-1].Hash, blockHeaders)
			network.Uptodate = true
		}
	}
}

func getState(acc *Account, lastTenTx []*FundsTxJson) (err error) {
	pubKeyHash := protocol.SerializeHashContent(acc.Address)

	//Get blocks if the Acc address:
	//* got issued as an Acc
	//* sent funds
	//* received funds
	//* is block's beneficiary
	//* nr of configTx in block is > 0 (in order to maintain params in light-client)
	relevantBlocks, err := getRelevantBlocks(acc.Address)

	for _, block := range relevantBlocks {
		if block != nil {
			//Collect block reward
			if block.Beneficiary == pubKeyHash {
				acc.Balance += activeParameters.Block_reward
			}

			//Balance funds and collect fee
			for _, txHash := range block.FundsTxData {
				err := network.TxReq(p2p.FUNDSTX_REQ, txHash)
				if err != nil {
					return err
				}

				txI, err := network.Fetch(network.FundsTxChan)
				if err != nil {
					return err
				}

				tx := txI.(protocol.Transaction)
				fundsTx := txI.(*protocol.FundsTx)

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
				err := network.TxReq(p2p.ACCTX_REQ, txHash)
				if err != nil {
					return err
				}

				txI, err := network.Fetch(network.AccTxChan)
				if err != nil {
					return err
				}

				tx := txI.(protocol.Transaction)
				accTx := txI.(*protocol.AccTx)

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
				err := network.TxReq(p2p.CONFIGTX_REQ, txHash)
				if err != nil {
					return err
				}

				txI, err := network.Fetch(network.ConfigTxChan)
				if err != nil {
					return err
				}

				tx := txI.(protocol.Transaction)
				configTx := txI.(*protocol.ConfigTx)

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
