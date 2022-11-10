package http

import (
	"context"
	nethttp "net/http"
	"runtime"

	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/faucet/app"
	"github.com/CoreumFoundation/faucet/pkg/http"
)

// HTTP type exposes app functionalities via http
type HTTP struct {
	app    app.App
	server http.Server
}

// New returns an instance of the HTTP type
func New(app app.App, log *zap.Logger) HTTP {
	return HTTP{
		app:    app,
		server: http.New(log, writeErrorMiddleware()),
	}
}

// ListenAndServe starts listening for http requests
func (h HTTP) ListenAndServe(ctx context.Context, address string) error {
	apiv1 := h.server.Group(
		"/api/faucet/v1",
		middleware.BodyLimit("4MB"),
	)

	apiv1.GET("/status", h.statusHandle)
	apiv1.POST("/fund", h.fundHandle)
	apiv1.POST("/gen-funded", h.genFundedHandle)
	return h.server.Start(ctx, address, 0)
}

// StatusResponse is the output to /status request
type StatusResponse struct {
	Version string `json:"version"`
	Status  string `json:"status"`
	Go      string `json:"go"`
}

func (h HTTP) statusHandle(ctx http.Context) error {
	return ctx.JSON(nethttp.StatusOK, StatusResponse{
		Version: "v1.0.0",
		Status:  "listening",
		Go:      runtime.Version(),
	})
}

// FundRequest is the input to GiveFunds request
type FundRequest struct {
	Address string `json:"address"`
}

// FundResponse is the output to GiveFunds request
type FundResponse struct {
	TxHash string `json:"txHash"`
}

func (h HTTP) fundHandle(ctx http.Context) error {
	var rqBody FundRequest
	if err := ctx.Bind(&rqBody); err != nil {
		return err
	}

	txHash, err := h.app.GiveFunds(ctx.Request().Context(), rqBody.Address)
	if err != nil {
		return err
	}

	return ctx.JSON(nethttp.StatusOK, FundResponse{TxHash: txHash})
}

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
