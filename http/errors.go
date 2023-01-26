package http

import (
	"encoding/json"
	nethttp "net/http"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/faucet/app"
	"github.com/CoreumFoundation/faucet/pkg/http"
)

// ErrRateLimitExhausted is returned when rate limit is exhausted for an IP address.
var ErrRateLimitExhausted = errors.New("rate limit exhausted")

func writeErrorMiddleware() func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(c http.Context) error {
			err := next(c)
			if err != nil {
				logger.Get(c.Request().Context()).Error("Error processing request", zap.Error(err))
				err := mapError(err)
				return c.JSON(err.Status(), err)
			}
			return nil
		}
	}
}

// APIError provides a wrapper around errors which makes exposing errors to outside world simpler.
type APIError interface {
	// Satisfy error interface.
	error
	// Override default marshaling.
	json.Marshaler
	// Method to return HTTP status code.
	Status() int
}

type singleAPIError struct {
	kind    string
	message string
	status  int
}

func newSingleAPIError(kind, message string, status int) singleAPIError {
	return singleAPIError{
		kind:    kind,
		message: message,
		status:  status,
	}
}

func (err singleAPIError) Error() string {
	return err.message
}

func (err singleAPIError) Status() int {
	return err.status
}

func (err singleAPIError) MarshalJSON() ([]byte, error) {
	type errEntity struct {
		Message string `json:"message"`
		Kind    string `json:"kind"`
	}
	resp := struct {
		Type    string      `json:"type"`
		Content []errEntity `json:"content"`
	}{
		Type: "errors",
		Content: []errEntity{
			{Message: err.message, Kind: err.kind},
		},
	}

	return json.Marshal(resp)
}

func mapError(err error) APIError {
	errList := map[error]singleAPIError{
		app.ErrAddressPrefixUnsupported: newSingleAPIError("address.invalid", app.ErrAddressPrefixUnsupported.Error(), nethttp.StatusUnprocessableEntity),
		app.ErrInvalidAddressFormat:     newSingleAPIError("address.invalid", app.ErrInvalidAddressFormat.Error(), nethttp.StatusUnprocessableEntity),
		app.ErrUnableToTransferToken:    newSingleAPIError("server.internal_error", app.ErrUnableToTransferToken.Error(), nethttp.StatusInternalServerError),
		ErrRateLimitExhausted:           newSingleAPIError("server.rate_limit", ErrRateLimitExhausted.Error(), nethttp.StatusTooManyRequests),
	}

	for e, internalErr := range errList {
		if errors.Is(err, e) {
			return internalErr
		}
	}

	return newSingleAPIError("server.internal_error", "internal error", nethttp.StatusInternalServerError)
}
