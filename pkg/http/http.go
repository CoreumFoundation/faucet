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

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
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
func New(log *zap.Logger, m ...MiddlewareFunc) Server {
	limiter := NewWeightedWindowLimiter(2, time.Hour)

	e := echo.New()
	e.Logger.SetLevel(99)
	e.HideBanner = true
	e.HidePort = true
	e.Use(prepareRequestContextMiddleware(log))
	e.Use(m...)
	e.Use(limiterMiddleware(limiter))
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:     true,
		LogStatus:  true,
		LogHeaders: []string{echo.HeaderXForwardedFor, HeaderXRequestID},
		LogMethod:  true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logger.Get(c.Request().Context()).Info("Request", zap.Int("status", v.Status))
			return nil
		},
	}))
	return Server{
		Echo:    e,
		limiter: limiter,
	}
}

// Server exposes functionalities needed to run an http server
type Server struct {
	*echo.Echo

	limiter *WeightedWindowLimiter
}

// Start begins listening and serving http requests with graceful shut down. graceful shutdown signal should be
// passed to the function as input and should come from the signal package.
// NOTE: graceful shutdown does not handle websocket and other hijacked connections (because it relies on http.server#Shutdown)
func (s Server) Start(ctx context.Context, listenAddress string, forceShutdownTimeout time.Duration) error {
	log := logger.Get(ctx)

	// Start server
	exitListening := make(chan error, 2)
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return errors.Wrap(err, "unable to listen on address")
	}
	go func() {
		var err error
		defer func() {
			if rec := recover(); rec != nil {
				err = errors.Wrapf(err, "listen panicked: %s", rec)
			}
			exitListening <- err
		}()
		log.Info("Started listening for http connections", zap.String("address", listenAddress))
		if err = http.Serve(listener, s.Echo); err != nil && !errors.Is(err, http.ErrServerClosed) {
			err = errors.Wrap(err, "Error listening for connections")
		}
	}()
	go func() {
		var err error
		defer func() {
			if rec := recover(); rec != nil {
				err = errors.Wrapf(err, "limiter panicked: %s", rec)
			}
			exitListening <- err
		}()
		err = s.limiter.Run(ctx)
	}()

	select {
	case <-ctx.Done():
	case err = <-exitListening:
		return err
	}
	if forceShutdownTimeout == 0 {
		forceShutdownTimeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(reopenCtx(ctx), forceShutdownTimeout)
	defer cancel()

	log.Info("Starting graceful shutdown")
	if err := s.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "Error shutting down server")
	}

	log.Info("Server shutdown successfully")
	return nil
}

func reopenCtx(ctx context.Context) context.Context {
	return reopened{Context: ctx}
}

type reopened struct {
	//nolint:containedctx // this struct exists to wrap a context
	context.Context
}

func (reopened) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

func (reopened) Done() <-chan struct{} {
	return nil
}

func (reopened) Err() error {
	return nil
}
