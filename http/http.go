package http

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/faucet/app"
	"github.com/pkg/errors"
	"go.uber.org/zap"
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
func (h HTTP) ListenAndServe(ctx context.Context, address string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/faucet/send-money", h.sendMoneyHandle)
	log := logger.Get(ctx)
	mwMux := middleware{
		next: mux,
		log:  log,
	}

	server := http.Server{
		Handler: mwMux,
		Addr:    address,
	}

	exitChan := make(chan error, 1)
	go func() {
		log.Info("Started listening for http connections", zap.String("address", address))
		exitChan <- server.ListenAndServe()
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	termSig := <-sigChan
	log.Info("Termination signal received", zap.Stringer("signal", termSig))

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		server.Close()
		return errors.Errorf("graceful shutdown failed with error %s", err)
	}

	if err := <-exitChan; !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
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

func (h HTTP) sendMoneyHandle(resp http.ResponseWriter, req *http.Request) {
	log := logger.Get(req.Context())
	var rqBody SendMoneyRequest
	err := parseJSONReqBody(req, &rqBody)
	if err != nil {
		respondErr(log, resp, err)
		return
	}

	txHash, err := h.app.GiveFunds(req.Context(), rqBody.Address)
	if err != nil {
		respondErr(log, resp, err)
		return
	}

	writeJSON(log, resp, SendMoneyResponse{TxHash: txHash}, http.StatusOK)
}

func parseJSONReqBody(req *http.Request, i interface{}) error {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return errors.WithStack(err)
	}
	return errors.WithStack(json.Unmarshal(body, i))
}

func writeJSON(log *zap.Logger, resp http.ResponseWriter, msg interface{}, statusCode int) {
	resp.Header().Add("Content-Type", "application/json")
	resp.WriteHeader(statusCode)
	encode := json.NewEncoder(resp)
	if err := encode.Encode(msg); err != nil {
		log.Error("Error encoding json message", zap.Any("msg", msg), zap.Error(err))
	}
}

func respondErr(log *zap.Logger, w http.ResponseWriter, err error) {
	log.Error("got error", zap.Error(err))
	errList := map[error]int{
		app.ErrAddressPrefixUnsupported: http.StatusNotAcceptable,
		app.ErrInvalidAddressFormat:     http.StatusNotAcceptable,
		app.ErrUnableToTransferToken:    http.StatusInternalServerError,
	}

	for e, status := range errList {
		if errors.Is(err, e) {
			writeJSON(log, w, ErrorResponse{Msg: e.Error()}, status)
			return
		}
	}

	writeJSON(log, w, ErrorResponse{Msg: "unable to fullfil request"}, http.StatusInternalServerError)
}
