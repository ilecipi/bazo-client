package REST

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/bazo-blockchain/bazo-client/client"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"github.com/gorilla/mux"
	"math/big"
	"net/http"
	"strconv"
)

type JsonResponse struct {
	Code    int       `json:"code,omitempty"`
	Message string    `json:"message,omitempty"`
	Content []Content `json:"content,omitempty"`
}

type Content struct {
	Name   string      `json:"name,omitempty"`
	Detail interface{} `json:"detail,omitempty"`
}

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
	logger.Println("Incoming createAcc request")

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

func CreateConfigTxEndpoint(w http.ResponseWriter, req *http.Request) {
	logger.Println("Incoming createConfig request")

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
	logger.Println("Incoming createFunds request")

	params := mux.Vars(req)

	var fromPub [64]byte
	var toPub [64]byte

	header, _ := strconv.Atoi(params["header"])
	amount, _ := strconv.Atoi(params["amount"])
	fee, _ := strconv.Atoi(params["fee"])
	txCnt, _ := strconv.Atoi(params["txCnt"])

	fromPubInt, _ := new(big.Int).SetString(params["fromPub"], 16)
	copy(fromPub[:], fromPubInt.Bytes())

	toPubInt, _ := new(big.Int).SetString(params["toPub"], 16)
	copy(toPub[:], toPubInt.Bytes())

	tx := protocol.FundsTx{
		Header: byte(header),
		Amount: uint64(amount),
		Fee:    uint64(fee),
		TxCnt:  uint32(txCnt),
		From:   protocol.SerializeHashContent(fromPub),
		To:     protocol.SerializeHashContent(toPub),
	}

	txHash := tx.Hash()
	client.UnsignedFundsTx[txHash] = &tx
	logger.Printf("New unsigned tx: %x\n", txHash)

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

	logger.Printf("Incoming sendTx request for tx: %x", txHash)

	switch txType {
	case p2p.ACCTX_BRDCST:
		if tx := client.UnsignedAccTx[txHash]; tx != nil {
			tx.Sig = txSign
			err = client.SendTx(storage.BOOTSTRAP_SERVER, tx, p2p.ACCTX_BRDCST)

			//If tx was successful or not, delete it from map either way. A new tx creation is the only option to repeat.
			delete(client.UnsignedFundsTx, txHash)
		} else {
			SendJsonResponse(w, JsonResponse{http.StatusInternalServerError, fmt.Sprintf("No transaction with hash %x found to sign", txHash), nil})
			return
		}
	case p2p.CONFIGTX_BRDCST:
		if tx := client.UnsignedConfigTx[txHash]; tx != nil {
			tx.Sig = txSign
			err = client.SendTx(storage.BOOTSTRAP_SERVER, tx, p2p.CONFIGTX_BRDCST)

			//If tx was successful or not, delete it from map either way. A new tx creation is the only option to repeat.
			delete(client.UnsignedFundsTx, txHash)
		} else {
			SendJsonResponse(w, JsonResponse{http.StatusInternalServerError, fmt.Sprintf("No transaction with hash %x found to sign", txHash), nil})
			return
		}
	case p2p.FUNDSTX_BRDCST:
		if tx := client.UnsignedFundsTx[txHash]; tx != nil {
			if tx.Sig1 == [64]byte{} {
				tx.Sig1 = txSign
				err = client.SendTx(client.MULTISIG_SERVER, tx, p2p.FUNDSTX_BRDCST)
				if err != nil {
					delete(client.UnsignedFundsTx, txHash)
				}
			} else {
				tx.Sig2 = txSign
				err = client.SendTx(storage.BOOTSTRAP_SERVER, tx, p2p.FUNDSTX_BRDCST)
				delete(client.UnsignedFundsTx, txHash)
			}
		} else {
			logger.Printf("No transaction with hash %x found to sign\n", txHash)
			SendJsonResponse(w, JsonResponse{http.StatusInternalServerError, fmt.Sprintf("No transaction with hash %x found to sign", txHash), nil})
			return
		}
	}

	if err == nil {
		SendJsonResponse(w, JsonResponse{http.StatusOK, fmt.Sprintf("Transaction %x successfully sent to network.", txHash[:8]), nil})
	} else {
		logger.Printf("Sending tx failed: %v\n", err.Error())
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
