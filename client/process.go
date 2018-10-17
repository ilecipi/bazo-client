package client

import (
	"bufio"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"math/big"
	"os"
	"strconv"
	"strings"
)

const (
	ARGS_MSG = "Wrong number of arguments."
)

func parseAccTx(args []string) (tx protocol.Transaction, err error) {
	accTxUsage := "\nUsage: bazo_client accTx <header> <fee> <root> <new>"

	if len(args) != 4 {
		return nil, errors.New(fmt.Sprintf("%v%v", ARGS_MSG, accTxUsage))
	}

	header, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, accTxUsage))
	}

	fee, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, accTxUsage))
	}

	_, privKey, err := storage.ExtractECDSAKeyFromFile(args[2])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, accTxUsage))
	}

	if _, err = os.Stat(args[3]); !os.IsNotExist(err) {
		return nil, errors.New(fmt.Sprintf("Output file exists.%v", accTxUsage))
	}

	if len(args[3]) == 128 {
		var newAddress [64]byte
		newPubInt, _ := new(big.Int).SetString(args[3], 16)
		copy(newAddress[:], newPubInt.Bytes())

		tx, _, err = protocol.ConstrAccTx(byte(header), uint64(fee), newAddress, &privKey, nil, nil)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("%v%v", err, accTxUsage))
		}
	} else {
		var newKey *ecdsa.PrivateKey
		//Write the public key to the given textfile
		file, err := os.Create(args[3])
		if err != nil {
			return nil, errors.New(fmt.Sprintf("%v%v", err, accTxUsage))
		}

		tx, newKey, err = protocol.ConstrAccTx(byte(header), uint64(fee), [64]byte{}, &privKey, nil, nil)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("%v%v", err, accTxUsage))
		}

		_, err = file.WriteString(string(newKey.X.Text(16)) + "\n")
		_, err = file.WriteString(string(newKey.Y.Text(16)) + "\n")
		_, err = file.WriteString(string(newKey.D.Text(16)) + "\n")

		if err != nil {
			return nil, errors.New(fmt.Sprintf("Failed to write key to file%v", accTxUsage))
		}
	}

	fmt.Printf("chash: %x\n", tx.Hash())
	return tx, nil
}

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

	_, privKey, err := storage.ExtractECDSAKeyFromFile(args[5])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, configTxUsage))
	}

	tx, err = protocol.ConstrConfigTx(byte(header), uint8(id), uint64(payload), uint64(fee), uint8(txCnt), &privKey)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, configTxUsage))
	}

	return tx, nil
}

func parseFundsTx(args []string) (tx protocol.Transaction, err error) {
	fundsTxUsage := "\nUsage: bazo_client fundsTx <header> <amount> <fee> <txCnt> <from> <to> <multiSig>"

	if len(args) != 7 {
		return nil, errors.New(fmt.Sprintf("%v%v", ARGS_MSG, fundsTxUsage))
	}

	header, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, fundsTxUsage))
	}

	amount, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, fundsTxUsage))
	}

	fee, err := strconv.Atoi(args[2])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, fundsTxUsage))
	}

	txCnt, err := strconv.Atoi(args[3])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, fundsTxUsage))
	}

	fromPubKey, fromPrivKey, err := storage.ExtractECDSAKeyFromFile(args[4])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, fundsTxUsage))
	}

	toPubKey, _, err := storage.ExtractECDSAKeyFromFile(args[5])
	if err != nil {
		if len(args[5]) == 128 {
			runes := []rune(args[5])
			pub1 := string(runes[:64])
			pub2 := string(runes[64:])

			toPubKey, _ = storage.GetPubKeyFromString(pub1, pub2)
		} else {
			return nil, errors.New(fmt.Sprintf("%v%v", err, fundsTxUsage))
		}
	}

	_, multiSigPrivKey, err := storage.ExtractECDSAKeyFromFile(args[6])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, fundsTxUsage))
	}

	fromAddress := storage.GetAddressFromPubKey(&fromPubKey)
	toAddress := storage.GetAddressFromPubKey(&toPubKey)

	tx, err = protocol.ConstrFundsTx(byte(header), uint64(amount), uint64(fee), uint32(txCnt), protocol.SerializeHashContent(fromAddress), protocol.SerializeHashContent(toAddress), &fromPrivKey, &multiSigPrivKey, nil)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, fundsTxUsage))
	}

	return tx, nil
}

func parseStakeTx(args []string) (tx protocol.Transaction, err error) {
	stakeTxUsage := "\nUsage: bazo_client stakeTx <header> <fee> <isStaking> <account> <privKey>"

	var (
		accountPubKey [64]byte
		hashedSeed    [32]byte
	)

	if len(args) != 5 {
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

	//create new seed if node wants to stake
	//seed file cannot be simply overwritten since in case of a rollback
	//the validator must also be able to access an old seed
	if isStaking != 0 {
		//generate random seed and store it
		seed := protocol.CreateRandomSeed()

		//create the hash of the seed which will be included in the transaction
		hashedSeed = protocol.SerializeHashContent(seed)

		storage.AppendNewSeed(args[4]+"_seed.json", storage.SeedJson{fmt.Sprintf("%x", string(hashedSeed[:])), string(seed[:])})
	}

	hashFromFile, err := os.Open(args[3])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, stakeTxUsage))
	}

	reader := bufio.NewReader(hashFromFile)

	//We only need the public key
	pub1, err := reader.ReadString('\n')
	pub2, err2 := reader.ReadString('\n')
	if err != nil || err2 != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, stakeTxUsage))
	}

	pub1Int, _ := new(big.Int).SetString(strings.Split(pub1, "\n")[0], 16)
	pub2Int, _ := new(big.Int).SetString(strings.Split(pub2, "\n")[0], 16)
	copy(accountPubKey[0:32], pub1Int.Bytes())
	copy(accountPubKey[32:64], pub2Int.Bytes())

	_, privKey, err := storage.ExtractECDSAKeyFromFile(args[4])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, stakeTxUsage))
	}
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
		hashedSeed,
		protocol.SerializeHashContent(accountPubKey[:]),
		&privKey,
	)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, stakeTxUsage))
	}

	if tx == nil {
		return nil, errors.New(fmt.Sprintf("Transaction encoding failed.%v", stakeTxUsage))
	}

	return tx, nil
}
