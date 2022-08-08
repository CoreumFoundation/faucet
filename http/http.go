package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
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
func (h HTTP) ListenAndServe(ctx context.Context, port int) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/faucet/send-money", h.sendMoneyHandle)
	logMWMux := loggerMiddleware{
		next: mux,
		log:  logger.Get(ctx),
	}
	return http.ListenAndServe(fmt.Sprintf(":%d", port), logMWMux)
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
	var rqBody SendMoneyRequest
	err := parseJSONReqBody(req, &rqBody)
	if err != nil {
		fmt.Println("err read body", err)
		respondErr(resp, err)
		return
	}

	txHash, err := h.app.GiveFunds(req.Context(), rqBody.Address)
	if err != nil {
		fmt.Println("err giving funds", err)
		respondErr(resp, err)
		return
	}

	writeJSON(resp, SendMoneyResponse{TxHash: txHash}, http.StatusOK)
}

type js map[string]interface{}

func parseJSONReqBody(req *http.Request, i interface{}) error {
	decoder := json.NewDecoder(req.Body)
	defer req.Body.Close()
	return decoder.Decode(i)
}

func writeJSON(resp http.ResponseWriter, msg interface{}, statusCode int) {
	resp.Header().Add("Content-Type", "application/json")
	resp.WriteHeader(statusCode)
	encode := json.NewEncoder(resp)
	if err := encode.Encode(msg); err != nil {
	}
}

func respondErr(resp http.ResponseWriter, err error) {
	writeJSON(resp, js{
		"error": err.Error(),
		"msg":   "got error",
	}, http.StatusInternalServerError)
}
