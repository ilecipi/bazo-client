package client

import (
	"encoding/hex"
	"errors"
	"fmt"
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

func GetAccount(pubKey [64]byte) (*Account, error) {
	//Initialize new account with empty address
	acc := Account{pubKey, hex.EncodeToString(pubKey[:]), 0, 0, false, false}

	//Set default params
	activeParameters = miner.NewDefaultParameters()

	//If Acc is Root in the bazo network state, we do not check for accTx, else we check
	if rootAcc := reqRootAccFromHash(protocol.SerializeHashContent(acc.Address)); rootAcc != nil {
		acc.IsCreated, acc.IsRoot = true, true
	}

	err = getState(&acc)
	if err != nil {
		return &acc, errors.New(fmt.Sprintf("Could not calculate state of account %x: %v\n", acc.Address[:8], err))
	}

	if acc.IsCreated == false {
		return nil, errors.New(fmt.Sprintf("Account %x does not exist.\n", acc.Address[:8]))
	}

	return &acc, nil
}

func (acc Account) String() string {
	addressHash := protocol.SerializeHashContent(acc.Address)
	return fmt.Sprintf("Hash: %x, Address: %x, TxCnt: %v, Balance: %v, isCreated: %v, isRoot: %v", addressHash[:8], acc.Address[:8], acc.TxCnt, acc.Balance, acc.IsCreated, acc.IsRoot)
}
