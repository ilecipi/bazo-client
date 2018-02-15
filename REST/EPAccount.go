package REST

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/bazo-blockchain/bazo-client/client"
	"math/big"
	"net/http"
)

func GetAccountEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

	var pubKey [64]byte
	pubKeyInt, _ := new(big.Int).SetString(params["id"], 16)
	copy(pubKey[:], pubKeyInt.Bytes())

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
