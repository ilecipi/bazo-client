package REST

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/bazo-blockchain/bazo-client/client"
	"github.com/bazo-blockchain/bazo-client/network"
	"github.com/bazo-blockchain/bazo-client/util"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/ed25519"
	"math/big"
	"net/http"
	"strconv"
	"sync"
)

var mutex = &sync.Mutex{};

type JsonResponse struct {
	Code    int       `json:"code,omitempty"`
	Message string    `json:"message,omitempty"`
	Content []Content `json:"content,omitempty"`
}

type Content struct {
	Name   string      `json:"name,omitempty"`
	Detail interface{} `json:"detail,omitempty"`
}

type IoTData struct {
	DevId     string `json:"DevId"`
	PublicKey []int  `json:"PublicKey"`
	Data      []int  `json:"Data"`
	Signature []int  `json:"Signature"`
	TxCnt     int    `json:"TxCnt"`
}

type AccTxIoT struct {
	DevId     string `json:"DevId"`
	PublicKey []int  `json:"PublicKey"`
	Issuer    []int  `json:"Issuer"`
	Fee       int    `json:"Fee"`
	TxCnt     int    `json:"TxCnt"`
}

type FundsTxIoT struct {
	DevId      string `json:"DevId"`
	ToPubKey   []int  `json:"ToPubKey"`
	FromPubKey []int  `json:"FromPubKey"`
	Amount     int    `json:"Amount"`
	Fee        int    `json:"Fee"`
	TxCnt      int    `json:"TxCnt"`
}

const (
	PUB_KEY_LEN   = 32
	SIGNATURE_LEN = 64
)

func CreateAccTxEndpoint(w http.ResponseWriter, req *http.Request) {
	logger.Println("Incoming createAcc request")

	params := mux.Vars(req)

	header, _ := strconv.Atoi(params["header"])
	fee, _ := strconv.Atoi(params["fee"])

	tx := protocol.AccTx{
		Header: byte(header),
		Fee:    uint64(fee),
	}

	issuerInt, _ := new(big.Int).SetString(params["issuer"], 16)
	copy(tx.Issuer[:], issuerInt.Bytes())

	newAccAddress, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	copy(tx.PubKey[:32], newAccAddress.PublicKey.X.Bytes())
	copy(tx.PubKey[32:], newAccAddress.PublicKey.Y.Bytes())

	txHash := tx.Hash()
	client.UnsignedAccTx[txHash] = &tx

	var content []Content
	content = append(content, Content{"PubKey1", hex.EncodeToString(tx.PubKey[:32])})
	content = append(content, Content{"PubKey2", hex.EncodeToString(tx.PubKey[32:])})
	content = append(content, Content{"PrivKey", hex.EncodeToString(newAccAddress.D.Bytes())})
	content = append(content, Content{"TxHash", hex.EncodeToString(txHash[:])})

	SendJsonResponse(w, JsonResponse{http.StatusOK, "AccTx successfully created.", content})
}

func CreateAccTxEndpointWithPubKey(w http.ResponseWriter, req *http.Request) {
	//logger.Println("Incoming createAcc request")

	params := mux.Vars(req)

	header, _ := strconv.Atoi(params["header"])
	fee, _ := strconv.Atoi(params["fee"])

	tx := protocol.AccTx{
		Header: byte(header),
		Fee:    uint64(fee),
	}

	fromPubInt, _ := new(big.Int).SetString(params["pubKey"], 16)
	copy(tx.PubKey[:], fromPubInt.Bytes())
	issuerInt, _ := new(big.Int).SetString(params["issuer"], 16)
	copy(tx.Issuer[:], issuerInt.Bytes())

	txHash := tx.Hash()
	client.UnsignedAccTx[txHash] = &tx

	var content []Content
	content = append(content, Content{"TxHash", hex.EncodeToString(txHash[:])})
	SendJsonResponse(w, JsonResponse{http.StatusOK, "AccTx successfully created.", content})
}

func CreateAccTxEndpointWithPubKeyIoT(w http.ResponseWriter, req *http.Request) {
	//logger.Println("Incoming createAccIoT request")

	params := mux.Vars(req)

	header, _ := strconv.Atoi(params["header"])

	var err error
	var accTxIoT AccTxIoT
	err = json.NewDecoder(req.Body).Decode(&accTxIoT)

	fee := accTxIoT.Fee
	tx := protocol.AccTx{
		Header: byte(header),
		Fee:    uint64(fee),
	}

	if req.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if len(accTxIoT.PublicKey) != PUB_KEY_LEN || len(accTxIoT.Issuer) != 32 {
		//TODO: response to the client
		http.Error(w, err.Error(), 400)
		return
	}

	toPub := [PUB_KEY_LEN]byte{}
	for index := range accTxIoT.PublicKey {
		toPub[index] = byte(accTxIoT.PublicKey[index])
	}

	issuer := [PUB_KEY_LEN]byte{}
	for index := range accTxIoT.Issuer {
		issuer[index] = byte(accTxIoT.Issuer[index])
	}
	copy(tx.PubKey[:], toPub[:])
	copy(tx.Issuer[:], issuer[:])

	txHash := tx.Hash()
	mutex.Lock()
	client.UnsignedAccTx[txHash] = &tx
	mutex.Unlock()
	var content []Content
	content = append(content, Content{"TxHash", hex.EncodeToString(txHash[:])})
	SendJsonResponse(w, JsonResponse{http.StatusOK, "AccTx successfully created.", content})
}

func CreateFundsTxIoT(w http.ResponseWriter, req *http.Request) {
	//logger.Println("Incoming createFunds request")

	params := mux.Vars(req)

	var fundsTxIoT FundsTxIoT
	var err error
	if req.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	err = json.NewDecoder(req.Body).Decode(&fundsTxIoT)

	header, _ := strconv.Atoi(params["header"])
	tx := protocol.FundsTx{
		Header: byte(header),
		Fee:    uint64(fundsTxIoT.Fee),
		Amount: uint64(fundsTxIoT.Amount),
		TxCnt:  uint32(fundsTxIoT.TxCnt),
	}

	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if len(fundsTxIoT.FromPubKey) != PUB_KEY_LEN || len(fundsTxIoT.ToPubKey) != PUB_KEY_LEN {
		//TODO: response to the client
		http.Error(w, err.Error(), 400)
		return
	}

	toPub := [PUB_KEY_LEN]byte{}
	for index := range fundsTxIoT.ToPubKey {
		toPub[index] = byte(fundsTxIoT.ToPubKey[index])
	}

	fromPubKey := [PUB_KEY_LEN]byte{}
	for index := range fundsTxIoT.FromPubKey {
		fromPubKey[index] = byte(fundsTxIoT.FromPubKey[index])
	}
	copy(tx.To[:], toPub[:])
	copy(tx.From[:], fromPubKey[:])
	tx.To = protocol.SerializeHashContent(tx.To)
	tx.From = protocol.SerializeHashContent(tx.From)

	txHash := tx.Hash()
	mutex.Lock()
	client.UnsignedFundsTx[txHash] = &tx
	mutex.Unlock()
	var content []Content
	content = append(content, Content{"TxHash", hex.EncodeToString(txHash[:])})
	SendJsonResponse(w, JsonResponse{http.StatusOK, "AccTx successfully created.", content})
}

func CreateConfigTxEndpoint(w http.ResponseWriter, req *http.Request) {
	//logger.Println("Incoming createConfig request")

	params := mux.Vars(req)

	header, _ := strconv.Atoi(params["header"])
	id, _ := strconv.Atoi(params["id"])
	payload, _ := strconv.Atoi(params["payload"])
	fee, _ := strconv.Atoi(params["fee"])
	txCnt, _ := strconv.Atoi(params["txCnt"])

	tx := protocol.ConfigTx{
		Header:  byte(header),
		Id:      uint8(id),
		Payload: uint64(payload),
		Fee:     uint64(fee),
		TxCnt:   uint8(txCnt),
	}

	txHash := tx.Hash()
	client.UnsignedConfigTx[txHash] = &tx

	var content []Content
	content = append(content, Content{"TxHash", hex.EncodeToString(txHash[:])})
	SendJsonResponse(w, JsonResponse{http.StatusOK, "ConfigTx successfully created.", content})
}

func CreateFundsTxEndpoint(w http.ResponseWriter, req *http.Request) {
	//logger.Println("Incoming createFunds request")

	params := mux.Vars(req)

	var fromPub [32]byte
	var toPub [32]byte

	header, _ := strconv.Atoi(params["header"])
	amount, _ := strconv.Atoi(params["amount"])
	fee, _ := strconv.Atoi(params["fee"])
	txCnt, _ := strconv.Atoi(params["txCnt"])

	fromPubInt, _ := new(big.Int).SetString(params["fromPub"], 16)
	copy(fromPub[:], fromPubInt.Bytes())

	toPubInt, _ := new(big.Int).SetString(params["toPub"], 16)
	copy(toPub[:], toPubInt.Bytes())
	//fmt.Println(fromPub)
	//fmt.Println(toPubInt)

	tx := protocol.FundsTx{
		Header: byte(header),
		Amount: uint64(amount),
		Fee:    uint64(fee),
		TxCnt:  uint32(txCnt),
		From:   protocol.SerializeHashContent(fromPub),
		To:     protocol.SerializeHashContent(toPub),
	}
	//fmt.Println("FUNDS", tx)
	txHash := tx.Hash()
	client.UnsignedFundsTx[txHash] = &tx
	//logger.Printf("New unsigned tx: %x\n", txHash)

	var content []Content
	content = append(content, Content{"TxHash", hex.EncodeToString(txHash[:])})
	SendJsonResponse(w, JsonResponse{http.StatusOK, "FundsTx successfully created.", content})
}

func sendTxEndpoint(w http.ResponseWriter, req *http.Request, txType int) {
	params := mux.Vars(req)

	var txHash [32]byte
	var txSign [64]byte
	var err error

	txHashInt, _ := new(big.Int).SetString(params["txHash"], 16)
	copy(txHash[:], txHashInt.Bytes())
	txSignInt, _ := new(big.Int).SetString(params["txSign"], 16)
	copy(txSign[:], txSignInt.Bytes())
	//logger.Printf("Incoming sendTx request for tx: %x", txHash)

	switch txType {
	case p2p.ACCTX_BRDCST:
		//logger.Print("ACCTX")

		if tx := client.UnsignedAccTx[txHash]; tx != nil {
			tx.Sig = txSign
			err = network.SendTx(util.Config.BootstrapIpport, tx, p2p.ACCTX_BRDCST)

			//If tx was successful or not, delete it from map either way. A new tx creation is the only option to repeat.
			delete(client.UnsignedFundsTx, txHash)
		} else {
			SendJsonResponse(w, JsonResponse{http.StatusInternalServerError, fmt.Sprintf("No transaction with hash %x found to sign", txHash), nil})
			return
		}
	case p2p.CONFIGTX_BRDCST:
		logger.Print("CONFIGTX")

		if tx := client.UnsignedConfigTx[txHash]; tx != nil {
			tx.Sig = txSign
			err = network.SendTx(util.Config.BootstrapIpport, tx, p2p.CONFIGTX_BRDCST)

			//If tx was successful or not, delete it from map either way. A new tx creation is the only option to repeat.
			delete(client.UnsignedFundsTx, txHash)
		} else {
			SendJsonResponse(w, JsonResponse{http.StatusInternalServerError, fmt.Sprintf("No transaction with hash %x found to sign", txHash), nil})
			return
		}
	case p2p.FUNDSTX_BRDCST:
		//logger.Print("FUNDSTX")
		mutex.Lock()
		if tx := client.UnsignedFundsTx[txHash]; tx != nil {
			if tx.Sig == [64]byte{} {
				tx.Sig = txSign
				err = network.SendTx(util.Config.BootstrapIpport, tx, p2p.FUNDSTX_BRDCST)
				if err != nil {
					delete(client.UnsignedFundsTx, txHash)
				}
			} else {
				tx.Sig = txSign
				err = network.SendTx(util.Config.BootstrapIpport, tx, p2p.FUNDSTX_BRDCST)
				delete(client.UnsignedFundsTx, txHash)
			}
		} else {
			//logger.Printf("No transaction with hash %x found to sign\n", txHash)
			SendJsonResponse(w, JsonResponse{http.StatusInternalServerError, fmt.Sprintf("No transaction with hash %x found to sign", txHash), nil})
			mutex.Unlock()

			return
		}
		mutex.Unlock()

	case p2p.IOTTX_BRDCST:
		//logger.Print("IOTTX")
		if tx := client.UnsignedIoTTx[txHash]; tx != nil {
			if tx.Sig == [64]byte{} {
				tx.Sig = txSign
				err = network.SendTx(util.Config.MultisigIpport, tx, p2p.IOTTX_BRDCST)
				if err != nil {
					delete(client.UnsignedFundsTx, txHash)
				}
			} else {
				tx.Sig = txSign
				err = network.SendTx(util.Config.BootstrapIpport, tx, p2p.IOTTX_BRDCST)
				delete(client.UnsignedFundsTx, txHash)
			}
		} else {
			//logger.Printf("No IoT transaction with hash %x found to sign\n", txHash)
			SendJsonResponse(w, JsonResponse{http.StatusInternalServerError, fmt.Sprintf("No transaction with hash %x found to sign", txHash), nil})
			return
		}
	}
	if err == nil {
		SendJsonResponse(w, JsonResponse{http.StatusOK, fmt.Sprintf("Transaction %x successfully sent to network.", txHash[:8]), nil})
	} else {
		//logger.Printf("Sending tx failed: %v\n", err.Error())
		SendJsonResponse(w, JsonResponse{http.StatusInternalServerError, err.Error(), nil})
	}
}

func SendAccTxEndpoint(w http.ResponseWriter, req *http.Request) {
	sendTxEndpoint(w, req, p2p.ACCTX_BRDCST)
}

func SendConfigTxEndpoint(w http.ResponseWriter, req *http.Request) {
	sendTxEndpoint(w, req, p2p.CONFIGTX_BRDCST)
}

func SendFundsTxEndpoint(w http.ResponseWriter, req *http.Request) {
	sendTxEndpoint(w, req, p2p.FUNDSTX_BRDCST)
}
func SendIoTTxEndpoint(w http.ResponseWriter, req *http.Request) {
	//logger.Println("Incoming IoT transaction...")
	params := mux.Vars(req)

	header, _ := strconv.Atoi(params["header"])

	//DYNAMIC OR STATIC?
	fee := 1

	var iotData IoTData
	var err error
	if req.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	err = json.NewDecoder(req.Body).Decode(&iotData)
	txCnt := iotData.TxCnt
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if len(iotData.PublicKey) != PUB_KEY_LEN || len(iotData.Signature) != SIGNATURE_LEN {
		//TODO: response to the client
		http.Error(w, err.Error(), 400)
		return
	}

	fromPub := [PUB_KEY_LEN]byte{}
	for index := range iotData.PublicKey {
		fromPub[index] = byte(iotData.PublicKey[index])
	}

	data := make([]byte, len(iotData.Data))
	for index := range iotData.Data {
		data[index] = byte(iotData.Data[index])
	}

	signature := [SIGNATURE_LEN]byte{}
	for index := range iotData.Signature {
		signature[index] = byte(iotData.Signature[index])
	}

	//fmt.Println("[PublicKey] ->\t", fromPub)
	//fmt.Println("[Data] ->\t\t", data)
	//fmt.Println("[Signature] ->\t", signature)
	//fmt.Println("[DevID] ->\t\t", iotData.DevId)

	valid := ed25519.Verify(ed25519.PublicKey(fromPub[:]), data, signature[:])
	toPublicKey, _ := crypto.ExtractEDPublicKeyFromFile("WalletA.txt")
	toPub := crypto.GetAddressFromPubKeyED(toPublicKey)

	//Check the signature on the client side so that we cannot flood the network with already invalid transactions
	if valid {
		//fmt.Println(valid)

		IotTx := protocol.IotTx{
			Header: byte(header),
			TxCnt:  uint32(txCnt),
			From:   protocol.SerializeHashContent(fromPub),
			To:     protocol.SerializeHashContent(toPub),
			Sig:    signature,
			Data:   data,
			Fee:    uint64(fee),
		}

		txHash := IotTx.Hash()
		mutex.Lock()
		client.SignedIotTx[txHash] = &IotTx
		tx := client.SignedIotTx[txHash]
		mutex.Unlock()

		err = network.SendIotTx(util.Config.BootstrapIpport, tx, p2p.IOTTX_BRDCST)

		if err == nil {
			SendJsonResponse(w, JsonResponse{http.StatusOK, fmt.Sprintf("Transaction %x successfully sent to network.", txHash[:8]), nil})

			var content []Content
			content = append(content, Content{"TxHash", hex.EncodeToString(txHash[:])})
			SendJsonResponse(w, JsonResponse{http.StatusOK, "FundsTx successfully created.", content})
		} else {
			logger.Printf("Sending IotTx failed: %v\n", err.Error())
			SendJsonResponse(w, JsonResponse{http.StatusInternalServerError, err.Error(), nil})
		}
	} else {
		return
	}
}
