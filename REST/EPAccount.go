package REST

import (
	"fmt"
	"github.com/bazo-blockchain/bazo-client/client"
	"github.com/bazo-blockchain/bazo-client/network"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/gorilla/mux"
	"math/big"
	"net/http"
)

func GetAccountEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	param := params["id"]

	logger.Printf("Incoming acc request for id: %v", param)

	var address [32]byte
	var addressHash [32]byte

	pubKeyInt, _ := new(big.Int).SetString(param, 16)
	fmt.Println(param)
	if len(param) == 64 {
		copy(addressHash[:], pubKeyInt.Bytes())

		network.AccReq(false, addressHash)

		accI, _ := network.Fetch(network.AccChan)
		acc := accI.(*protocol.Account)

		address = acc.Address
	} else if len(param) == 128 {
		copy(address[:], pubKeyInt.Bytes())
		addressHash = protocol.SerializeHashContent(address)
	}
	acc, lastTenTx, err := client.GetAccount(addressHash)
	fmt.Println(acc, lastTenTx)
	if err != nil {
		SendJsonResponse(w, JsonResponse{http.StatusInternalServerError, err.Error(), nil})
	} else {
		var content []Content
		content = append(content, Content{"account", acc})

		for _, tx := range lastTenTx {
			if tx != nil {
				content = append(content, Content{"inbound", tx})
			}
		}

		SendJsonResponse(w, JsonResponse{http.StatusOK, "", content})
	}
}
