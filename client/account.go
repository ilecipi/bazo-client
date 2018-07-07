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
	Address       [64]byte `json:"-"`
	AddressString string   `json:"address"`
	Balance       uint64   `json:"balance"`
	TxCnt         uint32   `json:"txCnt"`
	IsCreated     bool     `json:"isCreated"`
	IsRoot        bool     `json:"isRoot"`
}

func GetAccount(address [64]byte) (*Account, []*FundsTxJson, error) {
	//Initialize new account with empty address
	acc := Account{address, hex.EncodeToString(address[:]), 0, 0, false, false}
	var lastTenTx = make([]*FundsTxJson, 10)

	//Set default params
	activeParameters = miner.NewDefaultParameters()

	//If Acc is Root in the bazo network state, we do not check for accTx, else we check
	network.AccReq(true, protocol.SerializeHashContent(acc.Address))

	rootAccI, _ := network.Fetch(network.AccChan)
	rootAcc := rootAccI.(*protocol.Account)
	if rootAcc != nil {
		acc.IsRoot = true
	}

	err := getState(&acc, lastTenTx)
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintf("Could not calculate state of account %x: %v\n", acc.Address[:8], err))
	}

	//No accTx exists for this account since it is the initial root account
	//Add the initial root's balance
	if acc.IsCreated == false && acc.IsRoot == true {
		acc.IsCreated = true
		//TODO Take balance from active param
		acc.Balance += 1000 //staking_min + 1
	}

	if acc.IsCreated == false {
		return nil, nil, errors.New(fmt.Sprintf("Account %x does not exist.\n", acc.Address[:8]))
	}

	return &acc, lastTenTx, nil
}

func (acc Account) String() string {
	addressHash := protocol.SerializeHashContent(acc.Address)
	return fmt.Sprintf("Hash: %x, Address: %x, TxCnt: %v, Balance: %v, isCreated: %v, isRoot: %v", addressHash[:8], acc.Address[:8], acc.TxCnt, acc.Balance, acc.IsCreated, acc.IsRoot)
}
