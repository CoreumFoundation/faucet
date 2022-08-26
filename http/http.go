package http

import (
	"context"

	"github.com/labstack/echo/v4/middleware"
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
func (h HTTP) ListenAndServe(ctx context.Context, address string) error {
	h.server.Use(writeErrorMiddleware(h.logger))
	h.server.Use(middleware.BodyLimit("4MB"))
	h.server.GET("/api/faucet/v1/status", h.statusHandle)
	h.server.POST("/api/faucet/v1/send-money", h.sendMoneyHandle)

	return h.server.Start(ctx, address, 0)
}

// StatusResponse is the output to GiveFunds request
type StatusResponse struct {
	Version string `json:"version"`
	Status  string `json:"status"`
}

func (h HTTP) statusHandle(ctx http.Context) error {
	return ctx.JSON(200, StatusResponse{
		Version: "v1",
		Status:  "listening",
	})
}

// SendMoneyRequest is the input to GiveFunds request
type SendMoneyRequest struct {
	Address string `json:"address"`
}

// SendMoneyResponse is the output to GiveFunds request
type SendMoneyResponse struct {
	TxHash string `json:"txHash"`
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
