package http

import (
	"context"

	"go.uber.org/zap"

	"github.com/CoreumFoundation/faucet/app"
	"github.com/CoreumFoundation/faucet/pkg/http"
)

// HTTP type exposes app functionalities via http
type HTTP struct {
	app    app.App
	server http.Server
	logger *zap.Logger
}

// New returns an instance of the HTTP type
func New(app app.App, logger *zap.Logger) HTTP {
	return HTTP{
		app:    app,
		logger: logger,
		server: http.New(logger),
	}
}

// ListenAndServe starts listening for http requests
func (h HTTP) ListenAndServe(ctx context.Context, address string, shutdownSignal <-chan struct{}) {
	h.server.Use(writeErrorMiddleware(h.logger))
	h.server.GET("/api/v1/faucet/send-money", h.sendMoneyHandle)
	h.server.Start(address, shutdownSignal, 0)
}

// SendMoneyRequest is the input to GiveFunds method
type SendMoneyRequest struct {
	Address string `json:"address"`
}

// SendMoneyResponse is the output to GiveFunds method
type SendMoneyResponse struct {
	TxHash string `json:"txHash"`
}

// ErrorResponse is the response given in case an error occurred.
type ErrorResponse struct {
	Msg string `json:"msg"`
}

func (h HTTP) sendMoneyHandle(ctx http.Context) error {
	var rqBody SendMoneyRequest
	if err := ctx.Bind(&rqBody); err != nil {
		return err
	}

	txHash, err := h.app.GiveFunds(ctx.Request().Context(), rqBody.Address)
	if err != nil {
		return err
	}

	return ctx.JSON(200, SendMoneyResponse{TxHash: txHash})
}
