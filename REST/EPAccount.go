package REST

import (
	"github.com/bazo-blockchain/bazo-client/client"
	"github.com/gorilla/mux"
	"math/big"
	"net/http"
)

func GetAccountEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

	param := params["id"]
	var address [64]byte
	var addressHash [32]byte

	pubKeyInt, _ := new(big.Int).SetString(param, 16)

	if len(param) == 64 {
		copy(addressHash[:], pubKeyInt.Bytes())
		acc := client.ReqAcc(addressHash)
		address = acc.Address
	} else if len(param) == 128 {
		copy(address[:], pubKeyInt.Bytes())
	}

	acc, err := client.GetAccount(address)
	if err != nil {
		SendJsonResponse(w, JsonResponse{http.StatusInternalServerError, err.Error(), nil})
	} else {
		SendJsonResponse(w, JsonResponse{http.StatusOK, "", acc})
	}
}
