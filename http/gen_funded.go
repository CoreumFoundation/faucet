package http

import (
	nethttp "net/http"

	"github.com/CoreumFoundation/faucet/pkg/http"
)

// GenFundedResponse is the output to GiveFunds request
type GenFundedResponse struct {
	TxHash   string `json:"txHash"`
	Mnemonic string `json:"mnemonic"`
	Address  string `json:"address"`
}

func (h HTTP) genFundedHandle(ctx http.Context) error {
	result, err := h.app.GenMnemonicAndFund(ctx.Request().Context())
	if err != nil {
		return err
	}

	return ctx.JSON(nethttp.StatusOK, GenFundedResponse(result))
}
