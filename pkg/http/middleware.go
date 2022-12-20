package http

import (
	"net"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
)

// Predefined Headers
const (
	HeaderXRequestID = "X-Request-Id"
)

func prepareRequestContextMiddleware(log *zap.Logger) func(HandlerFunc) HandlerFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			r := c.Request()
			userIP, err := IPFromRequest(r)
			if err != nil {
				return err
			}

			rid := r.Header.Get(HeaderXRequestID)
			if rid == "" {
				rid = uuid.New().String()
				r.Header.Set(HeaderXRequestID, rid)
			}
			c.Response().Header().Set(HeaderXRequestID, rid)

			logNew := log.With(zap.String("URI", r.RequestURI),
				zap.Stringer("userIP", userIP),
				zap.String("requestID", rid),
				zap.String("method", r.Method),
			)
			ctx := logger.WithLogger(c.Request().Context(), logNew)
			request := c.Request().WithContext(ctx)
			c.SetRequest(request)
			return next(c)
		}
	}
}

// IPFromRequest returns IP of the client sending http request
func IPFromRequest(r *http.Request) (i net.IP, err error) {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	xForwardedFor := r.Header[echo.HeaderXForwardedFor]
	if len(xForwardedFor) > 0 {
		addrs := strings.Split(xForwardedFor[len(xForwardedFor)-1], ",")
		if addr := strings.TrimSpace(addrs[len(addrs)-1]); addr != "" {
			host = addr
		}
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return nil, errors.Errorf("failed to parse %q as an IP address", host)
	}
	return ip, nil
}
