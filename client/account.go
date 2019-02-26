package client

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-client/network"
	"github.com/bazo-blockchain/bazo-miner/miner"
	"github.com/bazo-blockchain/bazo-miner/protocol"
)

type Account struct {
	Address       [32]byte `json:"-"`
	AddressString string   `json:"address"`
	Balance       uint64   `json:"balance"`
	TxCnt         uint32   `json:"txCnt"`
	IsCreated     bool     `json:"isCreated"`
	IsRoot        bool     `json:"isRoot"`
	IsStaking     bool     `json:"isStaking"`
}

func CheckAccount(address [32]byte) (*Account, []*FundsTxJson, error) {
	loadBlockHeaders()
	return GetAccount(address)
}

func GetAccount(address [32]byte) (*Account, []*FundsTxJson, error) {
	//Initialize new account with empty address
	account := Account{address, hex.EncodeToString(address[:]), 0, 0, false, false, false}

	//Set default params
	activeParameters = miner.NewDefaultParameters()

	network.AccReq(false, protocol.SerializeHashContent(account.Address))
	if accI, _ := network.Fetch(network.AccChan); accI != nil {
		if acc := accI.(*protocol.Account); acc != nil {
			account.IsCreated = true
			account.IsStaking = acc.IsStaking

			//If Acc is Root in the bazo network state, we do not check for accTx, else we check
			network.AccReq(true, protocol.SerializeHashContent(account.Address))
			if rootAccI, _ := network.Fetch(network.AccChan); rootAccI != nil {
				if rootAcc := rootAccI.(*protocol.Account); rootAcc != nil {
					account.IsRoot = true
				}
			}
		}
	}

	if account.IsCreated == false {
		return nil, nil, errors.New(fmt.Sprintf("Account %x does not exist.\n", account.Address[:8]))
	}

	//if account.IsStaking == true {
	//	return nil, nil, errors.New(fmt.Sprintf("Account %x is a validator account. Validator's state cannot be calculated at the moment. We are sorry.\n", account.Address[:8]))
	//}

	var lastTenTx = make([]*FundsTxJson, 10)
	err := getState(&account, lastTenTx)
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintf("Could not calculate state of account %x: %v\n", account.Address[:8], err))
	}

	//No accTx exists for this account since it is the initial root account
	//Add the initial root's balance
	//if account.IsCreated == false && account.IsRoot == true {
	//	account.IsCreated = true
	//}

	return &account, lastTenTx, nil
}

func (acc Account) String() string {
	addressHash := protocol.SerializeHashContent(acc.Address)
	return fmt.Sprintf("Hash: %x, Address: %x, TxCnt: %v, Balance: %v, isCreated: %v, isRoot: %v", addressHash[:8], acc.Address[:8], acc.TxCnt, acc.Balance, acc.IsCreated, acc.IsRoot)
}
