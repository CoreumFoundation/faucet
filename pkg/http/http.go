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
	"github.com/CoreumFoundation/coreum-tools/pkg/parallel"
)

// Re-export types from echo library for convenience, so the users will not need to import echo library.
type (
	// HandlerFunc aliases and re-exports echo types so the users of this package don't need to reach to echo package.
	HandlerFunc = echo.HandlerFunc
	// MiddlewareFunc aliases and re-exports echo types so the users of this package don't need to reach to echo package.
	MiddlewareFunc = echo.MiddlewareFunc
	// Route aliases and re-exports echo types so the users of this package don't need to reach to echo package.
	Route = echo.Route
	// Context aliases and re-exports echo types so the users of this package don't need to reach to echo package.
	Context = echo.Context
)

// New returns a server instance.
func New(log *zap.Logger, middlewares ...MiddlewareFunc) Server {
	e := echo.New()
	e.Logger.SetLevel(99)
	e.HideBanner = true
	e.HidePort = true
	e.Use(prepareRequestContextMiddleware(log))
	e.Use(middlewares...)
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
	return Server{Echo: e}
}

// Server exposes functionalities needed to run an http server.
type Server struct {
	*echo.Echo
}

// Start begins listening and serving http requests with graceful shut down. graceful shutdown signal should be
// passed to the function as input and should come from the signal package.
// NOTE: graceful shutdown does not handle websocket and other hijacked connections
// (because it relies on http.server#Shutdown).
func (s Server) Start(ctx context.Context, listenAddress string, forceShutdownTimeout time.Duration) error {
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return errors.Wrap(err, "unable to listen on address")
	}

	return parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {
		spawn("listen", parallel.Fail, func(ctx context.Context) error {
			return s.listen(ctx, listener)
		})
		spawn("shutdown", parallel.Fail, func(ctx context.Context) error {
			return s.shutdown(ctx, forceShutdownTimeout)
		})
		return nil
	})
}

func (s Server) listen(ctx context.Context, listener net.Listener) error {
	logger.Get(ctx).Info("Started listening for http connections", zap.Stringer("address", listener.Addr()))
	if err := http.Serve(listener, s.Echo); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return errors.Wrap(err, "error listening for connections")
	}
	return errors.WithStack(ctx.Err())
}

func (s Server) shutdown(ctx context.Context, forceShutdownTimeout time.Duration) error {
	<-ctx.Done()
	log := logger.Get(ctx)

	ctx, cancel := context.WithTimeout(NewReopenedCtx(ctx), forceShutdownTimeout)
	defer cancel()

	log.Info("Starting graceful shutdown")
	if err := s.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "error shutting down server")
	}

	log.Info("Server shutdown successfully")
	return errors.WithStack(ctx.Err())
}

// NewReopenedCtx returns a context that inherits all the values stored in the given
// parent context, but not tied to the parent's lifespan. The returned context
// has no deadline. Reopen can even be used on an already closed context, hence the name.
func NewReopenedCtx(ctx context.Context) context.Context {
	return reopenedCtx{Context: ctx}
}

type reopenedCtx struct {
	//nolint:containedctx // this struct exists to wrap a context
	context.Context
}

func (reopenedCtx) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

func (reopenedCtx) Done() <-chan struct{} {
	return nil
}

func (reopenedCtx) Err() error {
	return nil
}
