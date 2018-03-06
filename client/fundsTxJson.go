package client

import (
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"encoding/hex"
)

type FundsTxJson struct {
	Header byte   `json:"header"`
	Amount uint64 `json:"amount"`
	Fee    uint64 `json:"fee"`
	TxCnt  uint32 `json:"txCnt"`
	From   string `json:"from"`
	To     string `json:"to"`
	Sig1   string `json:"sig1"`
	Sig2   string `json:"sig2"`
	Status string `json:"status"`
}

func ConvertFundsTx(fundsTx *protocol.FundsTx, status string) (fundsTxJson *FundsTxJson) {
	return &FundsTxJson{
		fundsTx.Header,
		fundsTx.Amount,
		fundsTx.Fee,
		fundsTx.TxCnt,
		hex.EncodeToString(fundsTx.From[:32]),
		hex.EncodeToString(fundsTx.To[:32]),
		hex.EncodeToString(fundsTx.Sig1[:64]),
		hex.EncodeToString(fundsTx.Sig2[:64]),
		status,
	}
}
