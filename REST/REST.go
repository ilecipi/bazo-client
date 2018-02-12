package REST

import (
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"github.com/bazo-blockchain/bazo-client/client"
)

var (
	logger *log.Logger
)

func Init() {
	logger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	client.InitState()

	logger.Printf("%v\n\n", "Starting REST...")

	router := mux.NewRouter()
	getEndpoints(router)
	log.Fatal(http.ListenAndServe(":443", handlers.CORS()(router)))
}

func getEndpoints(router *mux.Router) {
	router.HandleFunc("/account/{id}", GetAccountEndpoint).Methods("GET")

	router.HandleFunc("/createAccTx/{header}/{fee}/{issuer}", CreateAccTxEndpoint).Methods("POST")
	router.HandleFunc("/sendAccTx/{txHash}/{txSign}", SendAccTxEndpoint).Methods("POST")

	router.HandleFunc("/createConfigTx/{header}/{id}/{payload}/{fee}/{txCnt}", CreateConfigTxEndpoint).Methods("POST")
	router.HandleFunc("/sendConfigTx/{txHash}/{txSign}", SendConfigTxEndpoint).Methods("POST")

	router.HandleFunc("/createFundsTx/{header}/{amount}/{fee}/{txCnt}/{fromPub}/{toPub}", CreateFundsTxEndpoint).Methods("POST")
	router.HandleFunc("/sendFundsTx/{txHash}/{txSign}", SendFundsTxEndpoint).Methods("POST")
}
