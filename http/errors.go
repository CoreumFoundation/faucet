package http

import (
	"encoding/json"
	nethttp "net/http"

	"github.com/labstack/echo/v4"
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
				// handle default echo errors such as 404
				var echoError *echo.HTTPError
				if errors.As(err, &echoError) {
					return err
				}
				mappedError := mapError(err)
				if mappedError.Loggable() {
					logger.Get(c.Request().Context()).Error("Error processing request", zap.Error(err))
				}

				return c.JSON(mappedError.Status(), mappedError)
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

	// Status method to return HTTP status code.
	Status() int

	// Loggable indicates whether we need to log that error.
	Loggable() bool
}

type singleAPIError struct {
	kind     string
	message  string
	status   int
	loggable bool
}

func newSingleAPIError(kind, message string, status int, loggable bool) singleAPIError {
	return singleAPIError{
		kind:     kind,
		message:  message,
		status:   status,
		loggable: loggable,
	}
}

func (err singleAPIError) Error() string {
	return err.message
}

func (err singleAPIError) Status() int {
	return err.status
}

func (err singleAPIError) Loggable() bool {
	return err.loggable
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
		app.ErrAddressPrefixUnsupported: newSingleAPIError("address.invalid", app.ErrAddressPrefixUnsupported.Error(), nethttp.StatusUnprocessableEntity, false),
		app.ErrInvalidAddressFormat:     newSingleAPIError("address.invalid", app.ErrInvalidAddressFormat.Error(), nethttp.StatusUnprocessableEntity, false),
		app.ErrUnableToTransferToken:    newSingleAPIError("server.internal_error", app.ErrUnableToTransferToken.Error(), nethttp.StatusInternalServerError, true),
		ErrRateLimitExhausted:           newSingleAPIError("server.rate_limit", ErrRateLimitExhausted.Error(), nethttp.StatusTooManyRequests, false),
	}

	for e, internalErr := range errList {
		if errors.Is(err, e) {
			return internalErr
		}
	}

	return newSingleAPIError("server.internal_error", "internal error", nethttp.StatusInternalServerError, true)
}
