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

	var pubKey [64]byte
	var pubKeyHash [32]byte

	idInt, _ := new(big.Int).SetString(params["id"], 16)

	if len(params["id"]) == 64 {
		copy(pubKeyHash[:], idInt.Bytes())
		acc := client.ReqAcc(pubKeyHash)
		pubKey = acc.Address
	} else if len(params["id"]) == 128 {
		copy(pubKey[:], idInt.Bytes())
	}

	acc, err := client.GetAccount(pubKey)
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
