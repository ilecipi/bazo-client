package REST

import (
	"encoding/json"
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

	pubKeyInt, _ := new(big.Int).SetString(params["id"], 16)

	if len(param) == 64 {
		copy(addressHash[:], pubKeyInt.Bytes())
		acc := client.ReqAcc(addressHash)
		address = acc.Address
	} else if len(param) == 128 {
		copy(address[:], pubKeyInt.Bytes())
	}

	acc, err := client.GetAccount(address)
	if err != nil {
		js, err := json.Marshal(err.Error())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	} else {
		js, err := json.Marshal(acc)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}
