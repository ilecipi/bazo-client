package client

import (
	"encoding/hex"
	"github.com/bazo-blockchain/bazo-miner/protocol"
)

type FundsTxJson struct {
	Header byte   `json:"header"`
	Hash   string `json:"hash"`
	Amount uint64 `json:"amount"`
	Fee    uint64 `json:"fee"`
	TxCnt  uint32 `json:"txCnt"`
	From   string `json:"from"`
	To     string `json:"to"`
	Sig   string `json:"sig"`
	Status string `json:"status"`
}

func ConvertFundsTx(fundsTx *protocol.FundsTx, status string) (fundsTxJson *FundsTxJson) {
	txHash := fundsTx.Hash()
	return &FundsTxJson{
		fundsTx.Header,
		hex.EncodeToString(txHash[:]),
		fundsTx.Amount,
		fundsTx.Fee,
		fundsTx.TxCnt,
		hex.EncodeToString(fundsTx.From[:]),
		hex.EncodeToString(fundsTx.To[:]),
		hex.EncodeToString(fundsTx.Sig[:]),
		status,
	}
}
