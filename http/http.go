package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/CoreumFoundation/faucet/app"
)

// HTTP type exposes app functionalities via http
type HTTP struct {
	app app.App
}

// New returns an instance of the HTTP type
func New(app app.App) HTTP {
	return HTTP{
		app: app,
	}
}

// ListenAndServe starts listening for http requests
func (h HTTP) ListenAndServe(ctx context.Context, port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/faucet/send-money", h.sendMoneyHandle)
	http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

// SendMoneyRequest is the input to GiveFunds method
type SendMoneyRequest struct {
	Address string `json:"address"`
}

// SendMoneyResponse is the output to GiveFunds method
type SendMoneyResponse struct {
	TxHash string `json:"txHash"`
}

func (h HTTP) sendMoneyHandle(resp http.ResponseWriter, req *http.Request) {

}
