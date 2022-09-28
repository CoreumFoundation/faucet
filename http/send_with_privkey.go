package http

import "github.com/CoreumFoundation/faucet/pkg/http"

// SendMoneyGenPrivkeyResponse is the output to GiveFunds request
type SendMoneyGenPrivkeyResponse struct {
	TxHash   string `json:"txHash"`
	Mnemonic string `json:"mnemonic"`
	Address  string `json:"address"`
}

func (h HTTP) sendMoneyGenPrivkeyHandle(ctx http.Context) error {
	rsp, err := h.app.GenPrivkeyAndFund(ctx.Request().Context())
	if err != nil {
		return err
	}

	return ctx.JSON(200, SendMoneyGenPrivkeyResponse{
		TxHash:   rsp.TxHash,
		Mnemonic: rsp.Mnemonic,
		Address:  rsp.Address,
	})
}
