package http

import (
	"encoding/json"
	"errors"
	stdHttp "net/http"

	"go.uber.org/zap"

	"github.com/CoreumFoundation/faucet/app"
	"github.com/CoreumFoundation/faucet/pkg/http"
)

func withEchoContext(logger *zap.Logger, c http.Context) *zap.Logger {
	return logger.With(
		zap.String("request_id", c.Request().Header.Get(http.HeaderXRequestID)),
	)
}

func writeErrorMiddleware(logger *zap.Logger) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(c http.Context) error {
			err := next(c)
			if err != nil {
				logger = withEchoContext(logger, c)
				err := mapError(err)
				logger.Error("error processing request", zap.Error(err))
				return c.JSON(err.Status(), err)
			}
			return nil
		}
	}
}

// APIError provides a wrapper around errors which makes exposing errors to outside world simpler
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
			{Message: err.message},
		},
	}

	return json.Marshal(resp)
}

func mapError(err error) APIError {
	errList := map[error]singleAPIError{
		app.ErrAddressPrefixUnsupported: newSingleAPIError("address.invalid", app.ErrAddressPrefixUnsupported.Error(), stdHttp.StatusNotAcceptable),
		app.ErrInvalidAddressFormat:     newSingleAPIError("address.invalid", app.ErrInvalidAddressFormat.Error(), stdHttp.StatusNotAcceptable),
		app.ErrUnableToTransferToken:    newSingleAPIError("server.internal_error", app.ErrUnableToTransferToken.Error(), stdHttp.StatusInternalServerError),
	}

	for e, internalErr := range errList {
		if errors.Is(err, e) {
			return internalErr
		}
	}

	return newSingleAPIError("server.internal_error", "internal error", stdHttp.StatusInternalServerError)
}
