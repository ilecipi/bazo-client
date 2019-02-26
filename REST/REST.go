package REST

import (
	"encoding/json"
	"github.com/bazo-blockchain/bazo-client/util"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

var (
	logger *log.Logger
)

func Init() {
	logger = util.InitLogger()

	logger.Printf("%v\n\n", "Starting REST...")

	router := mux.NewRouter()
	getEndpoints(router)
	log.Fatal(http.ListenAndServe(":"+util.Config.Thisclient.Port, handlers.CORS()(router)))
}

func getEndpoints(router *mux.Router) {
	router.HandleFunc("/account/{id}", GetAccountEndpoint).Methods("GET")

	router.HandleFunc("/createAccTx/{header}/{fee}/{issuer}", CreateAccTxEndpoint).Methods("POST")
	router.HandleFunc("/createAccTx/{pubKey}/{header}/{fee}/{issuer}", CreateAccTxEndpointWithPubKey).Methods("POST")
	router.HandleFunc("/sendAccTx/{txHash}/{txSign}", SendAccTxEndpoint).Methods("POST")

	router.HandleFunc("/createConfigTx/{header}/{id}/{payload}/{fee}/{txCnt}", CreateConfigTxEndpoint).Methods("POST")
	router.HandleFunc("/sendConfigTx/{txHash}/{txSign}", SendConfigTxEndpoint).Methods("POST")

	router.HandleFunc("/createFundsTx/{header}/{amount}/{fee}/{txCnt}/{fromPub}/{toPub}", CreateFundsTxEndpoint).Methods("POST")
	router.HandleFunc("/sendFundsTx/{txHash}/{txSign}", SendFundsTxEndpoint).Methods("POST")
	router.HandleFunc("/verify", VerifyData).Methods("POST")
}

func SendJsonResponse(w http.ResponseWriter, resp interface{}) {
	js, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
