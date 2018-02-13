package REST

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/bazo-blockchain/bazo-client/client"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"math/big"
	"net/http"
	"strconv"
)

type JsonResponse struct {
	Code    int         `json:"code,omitempty"`
	Message string      `json:"message,omitempty"`
	Content interface{} `json:"content,omitempty"`
}

type Content struct {
	Name   string `json:"name,omitempty"`
	Detail string `json:"detail,omitempty"`
}

func CreateAccTxEndpoint(w http.ResponseWriter, req *http.Request) {
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

	var content [4]Content
	content[0] = Content{"PubKey1", hex.EncodeToString(tx.PubKey[:32])}
	content[1] = Content{"PubKey2", hex.EncodeToString(tx.PubKey[32:])}
	content[2] = Content{"PrivKey", hex.EncodeToString(newAccAddress.D.Bytes())}
	content[3] = Content{"TxHash", hex.EncodeToString(txHash[:])}

	sendJsonResponse(w, JsonResponse{200, "AccTx successfully created.", content})
}

func CreateConfigTxEndpoint(w http.ResponseWriter, req *http.Request) {
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

	var content [1]Content
	content[0] = Content{"TxHash", hex.EncodeToString(txHash[:])}
	sendJsonResponse(w, JsonResponse{200, "ConfigTx successfully created.", content})
}

func CreateFundsTxEndpoint(w http.ResponseWriter, req *http.Request) {
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

	var content [1]Content
	content[0] = Content{"TxHash", hex.EncodeToString(txHash[:])}
	sendJsonResponse(w, JsonResponse{200, "FundsTx successfully created.", content})
}

func sendTxEndpoint(w http.ResponseWriter, req *http.Request, txType int) {
	params := mux.Vars(req)

	var txHash [32]byte
	var txSign [64]byte

	txHashInt, _ := new(big.Int).SetString(params["txHash"], 16)
	copy(txHash[:], txHashInt.Bytes())
	txSignInt, _ := new(big.Int).SetString(params["txSign"], 16)
	copy(txSign[:], txSignInt.Bytes())

	var err error

	switch txType {
	case p2p.ACCTX_BRDCST:
		if tx := client.UnsignedAccTx[txHash]; tx != nil {
			tx.Sig = txSign
			if err = client.SendTx(p2p.BOOTSTRAP_SERVER, tx, p2p.ACCTX_BRDCST); err != nil {
				delete(client.UnsignedAccTx, txHash)
			}
		} else {
			sendJsonResponse(w, JsonResponse{500, fmt.Sprintf("No transaction with hash %x found to sign", txHash), nil})
			return
		}
	case p2p.CONFIGTX_BRDCST:
		if tx := client.UnsignedConfigTx[txHash]; tx != nil {
			tx.Sig = txSign
			if err = client.SendTx(p2p.BOOTSTRAP_SERVER, tx, p2p.CONFIGTX_BRDCST); err != nil {
				delete(client.UnsignedConfigTx, txHash)
			}
		} else {
			sendJsonResponse(w, JsonResponse{500, fmt.Sprintf("No transaction with hash %x found to sign", txHash), nil})
			return
		}
	case p2p.FUNDSTX_BRDCST:
		if tx := client.UnsignedFundsTx[txHash]; tx != nil {
			tx.Sig = txSign
			if err = client.SendTx(p2p.BOOTSTRAP_SERVER, tx, p2p.FUNDSTX_BRDCST); err != nil {
				delete(client.UnsignedFundsTx, txHash)
			}
		} else {
			sendJsonResponse(w, JsonResponse{500, fmt.Sprintf("No transaction with hash %x found to sign", txHash), nil})
			return
		}
	}

	if err == nil {
		sendJsonResponse(w, JsonResponse{200, fmt.Sprintf("Transaction successfully sent to network: %x", txHash), nil})
	} else {
		sendJsonResponse(w, JsonResponse{500, err.Error(), nil})
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

func sendJsonResponse(w http.ResponseWriter, resp interface{}) {
	js, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
