package client

import (
	"bufio"
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

func parseAccTx(args []string) (protocol.Transaction, error) {
	accTxUsage := "\nUsage: bazo_client accTx <header> <fee> <privKey> <keyOutput>"

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

	_, privKey, err := ExtractKeyFromFile(args[2])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, accTxUsage))
	}

	tx, newKey, err := protocol.ConstrAccTx(byte(header), uint64(fee), &privKey)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, accTxUsage))
	}

	if tx == nil {
		return nil, errors.New(fmt.Sprintf("Transaction encoding failed.%v", accTxUsage))
	}

	//Write the public key to the given textfile
	if _, err = os.Stat(args[3]); !os.IsNotExist(err) {
		return nil, errors.New(fmt.Sprintf("Output file exists.%v", accTxUsage))
	}

	file, err := os.Create(args[3])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, accTxUsage))
	}

	_, err = file.WriteString(string(newKey.X.Text(16)) + "\n")
	_, err2 := file.WriteString(string(newKey.Y.Text(16)) + "\n")
	_, err3 := file.WriteString(string(newKey.D.Text(16)) + "\n")

	if err != nil || err2 != nil || err3 != nil {
		return nil, errors.New(fmt.Sprintf("Failed to write key to file%v", accTxUsage))
	}

	return tx, nil
}

func parseConfigTx(args []string) (protocol.Transaction, error) {
	options := "\nOptions: <id> <payload [format]>\n 1 block size [bytes]\n 2 difficulty interval [#blocks]\n 3 minimum fee [bazo coins]\n 4 block interval [sec]\n 5 block reward [bazo coins]"
	configTxUsage := "\nUsage: bazo_client configTx <header> <id> <payload> <fee> <txCnt> <privKey>" + options

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

	_, privKey, err := ExtractKeyFromFile(args[5])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, configTxUsage))
	}

	tx, err := protocol.ConstrConfigTx(
		byte(header),
		uint8(id),
		uint64(payload),
		uint64(fee),
		uint8(txCnt),
		&privKey,
	)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, configTxUsage))
	}

	if tx == nil {
		return nil, errors.New(fmt.Sprintf("Transaction encoding failed.%v", configTxUsage))
	}

	return tx, nil
}

func parseFundsTx(args []string) (protocol.Transaction, error) {
	fundsTxUsage := "\nUsage: bazo_client fundsTx <header> <amount> <fee> <txCnt> <fromHash> <toHash> <privKey>"

	var (
		fromPubKey, toPubKey [64]byte
	)

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

	hashFromFile, err := os.Open(args[4])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, fundsTxUsage))
	}

	reader := bufio.NewReader(hashFromFile)
	//We only need the public key
	pub1, err := reader.ReadString('\n')
	pub2, err2 := reader.ReadString('\n')
	if err != nil || err2 != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, fundsTxUsage))
	}

	pub1Int, _ := new(big.Int).SetString(strings.Split(pub1, "\n")[0], 16)
	pub2Int, _ := new(big.Int).SetString(strings.Split(pub2, "\n")[0], 16)
	copy(fromPubKey[0:32], pub1Int.Bytes())
	copy(fromPubKey[32:64], pub2Int.Bytes())

	hashToFile, err := os.Open(args[5])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, fundsTxUsage))
	}

	reader.Reset(hashToFile)
	//We only need the public key
	pub1, err = reader.ReadString('\n')
	pub2, err2 = reader.ReadString('\n')
	if err != nil || err2 != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, fundsTxUsage))
	}

	pub1Int, _ = new(big.Int).SetString(strings.Split(pub1, "\n")[0], 16)
	pub2Int, _ = new(big.Int).SetString(strings.Split(pub2, "\n")[0], 16)
	copy(toPubKey[0:32], pub1Int.Bytes())
	copy(toPubKey[32:64], pub2Int.Bytes())

	_, privKey, err := ExtractKeyFromFile(args[6])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, fundsTxUsage))
	}

	tx, err := protocol.ConstrFundsTx(
		byte(header),
		uint64(amount),
		uint64(fee),
		uint32(txCnt),
		protocol.SerializeHashContent(fromPubKey[:]),
		protocol.SerializeHashContent(toPubKey[:]),
		&privKey,
	)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, fundsTxUsage))
	}

	if tx == nil {
		return nil, errors.New(fmt.Sprintf("Transaction encoding failed.%v", fundsTxUsage))
	}

	return tx, nil
}

func parseStakeTx(args []string) (protocol.Transaction, error) {
	stakeTxUsage := "\nUsage: bazo_client stakeTx <header> <fee> <isStaking> <account> <privKey>"

	var (
		accountPubKey [64]byte
		hashedSeed    [32]byte
	)

	if len(args) != 5 {
		return nil, errors.New(fmt.Sprintf("%v%v", ARGS_MSG, stakeTxUsage))
	}

	//TODO not needed!
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
		hashedSeed = SerializeHashContent(seed)

		storage.AppendNewSeed(storage.SEED_FILE_NAME, storage.SeedJson{fmt.Sprintf("%x", string(hashedSeed[:])), string(seed[:])})

		logger.Printf("%x", string(hashedSeed[:]))
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

	_, privKey, err := ExtractKeyFromFile(args[4])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, stakeTxUsage))
	}

	//logger.Println("\n Pubkey from ParseStakeTx: ", accountPubKey[:])
	//logger.Println("\nHashed Pubkey from ParseStakeTx: ", SerializeHashContent(accountPubKey[:]))

	tx, err := protocol.ConstrStakeTx(
		byte(header),
		uint64(fee),
		isStaking != 0,
		hashedSeed,
		SerializeHashContent(accountPubKey[:]),
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
