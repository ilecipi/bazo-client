package client

import (
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"strconv"
)

const (
	ARGS_MSG = "Wrong number of arguments."
)

func parseConfigTx(args []string) (tx protocol.Transaction, err error) {
	//TODO add new options
	options := "\nOptions: <id> <payload [format]>\n 1 block size [bytes]\n 2 difficulty interval [#blocks]\n 3 minimum fee [bazo coins]\n 4 block interval [sec]\n 5 block reward [bazo coins]"
	configTxUsage := "\nUsage: bazo_client configTx <header> <id> <payload> <fee> <txCnt> <root>" + options

	if len(args) != 6 {
		return nil, errors.New(fmt.Sprintf("%v%v", ARGS_MSG, configTxUsage))
	}

	header, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, configTxUsage))
	}

	id, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, configTxUsage))
	}

	payload, err := strconv.Atoi(args[2])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, configTxUsage))
	}

	fee, err := strconv.Atoi(args[3])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, configTxUsage))
	}

	txCnt, err := strconv.Atoi(args[4])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, configTxUsage))
	}

	privKey, err := crypto.ExtractECDSAKeyFromFile(args[5])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, configTxUsage))
	}

	tx, err = protocol.ConstrConfigTx(byte(header), uint8(id), uint64(payload), uint64(fee), uint8(txCnt), privKey)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, configTxUsage))
	}

	return tx, nil
}

func parseStakeTx(args []string) (tx protocol.Transaction, err error) {
	stakeTxUsage := "\nUsage: bazo_client stakeTx <header> <fee> <isStaking> <account> <privKey> <commitmentFile>"

	if len(args) != 6 {
		return nil, errors.New(fmt.Sprintf("%v%v", ARGS_MSG, stakeTxUsage))
	}

	header, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, stakeTxUsage))
	}

	fee, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, stakeTxUsage))
	}

	isStaking, err := strconv.Atoi(args[2])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, stakeTxUsage))
	}

	pubKey, err := crypto.ExtractECDSAPublicKeyFromFile(args[3])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, stakeTxUsage))
	}
	accountPubKey := crypto.GetAddressFromPubKey(pubKey)

	privKey, err := crypto.ExtractECDSAKeyFromFile(args[4])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, stakeTxUsage))
	}

	commPrivKey, err := crypto.ExtractRSAKeyFromFile(args[5])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, stakeTxUsage))
	}

	var isStakingAsBool bool
	if isStaking == 0 {
		isStakingAsBool = false
	} else {
		isStakingAsBool = true
	}

	tx, err = protocol.ConstrStakeTx(
		byte(header),
		uint64(fee),
		isStakingAsBool,
		protocol.SerializeHashContent(accountPubKey),
		privKey,
		&commPrivKey.PublicKey,
	)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, stakeTxUsage))
	}

	if tx == nil {
		return nil, errors.New(fmt.Sprintf("Transaction encoding failed.%v", stakeTxUsage))
	}

	return tx, nil
}
