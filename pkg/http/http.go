package http

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// re-export types from echo library for convenience, so the users will not need to import echo library
type (
	// HandlerFunc aliases and re-exports echo types so the users of this package don't need to reach to echo package
	HandlerFunc = echo.HandlerFunc
	// MiddlewareFunc aliases and re-exports echo types so the users of this package don't need to reach to echo package
	MiddlewareFunc = echo.MiddlewareFunc
	// Route aliases and re-exports echo types so the users of this package don't need to reach to echo package
	Route = echo.Route
	// Context aliases and re-exports echo types so the users of this package don't need to reach to echo package
	Context = echo.Context
)

// New returns a server instance
func New(logger *zap.Logger) Server {
	e := echo.New()
	e.Logger.SetLevel(99)
	e.HideBanner = true
	e.HidePort = true
	e.Use(addLoggerToRequestContext(logger))
	e.Use(requestIDMiddleware)
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:     true,
		LogStatus:  true,
		LogHeaders: []string{echo.HeaderXForwardedFor, HeaderXRequestID},
		LogMethod:  true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logger.Info("request",
				zap.String("URI", v.URI),
				zap.Int("status", v.Status),
				zap.Strings("user_real_ip", v.Headers[echo.HeaderXForwardedFor]),
				zap.String("request_id", v.Headers[HeaderXRequestID][0]),
				zap.String("method", v.Method),
			)

			return nil
		},
	}))
	return Server{Echo: e, logger: logger}
}

// Server exposes functionalities needed to run an http server
type Server struct {
	*echo.Echo
	logger *zap.Logger
}

// Start begins listening and serving http requests with graceful shut down. graceful shutdown signal should be
// passed to the function as input and should come from the signal package.
// NOTE: graceful shutdown does not handle websocket and other hijacked connections (because it relies on http.server#Shutdown)
func (s Server) Start(ctx context.Context, listenAddress string, forceShutdownTimeout time.Duration) error {
	// Start server
	exitListening := make(chan error, 1)
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return errors.Wrap(err, "unable to listen on address")
	}
	go func() {
		var err error
		defer func() {
			if rec := recover(); rec != nil {
				err = errors.Wrapf(err, "listen paniced %s", rec)
				s.logger.Error("listen paniced", zap.Error(err))
			}
			exitListening <- err
		}()
		s.logger.Info("Started listening for http connections", zap.String("address", listenAddress))
		if err = http.Serve(listener, s.Echo); err != nil && !errors.Is(err, http.ErrServerClosed) {
			err = errors.Wrap(err, "Error listening for connections")
		}
	}()

	select {
	case <-ctx.Done():
	case err = <-exitListening:
		return err
	}
	if forceShutdownTimeout == 0 {
		forceShutdownTimeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), forceShutdownTimeout)
	defer cancel()

	s.logger.Info("Starting graceful shutdown")
	if err := s.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "Error shutting down server")
	}

	s.logger.Info("Server shutdown successfully")
	return nil
}
