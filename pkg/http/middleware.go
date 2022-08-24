package http

import (
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
)

// Predefined Headers
const (
	HeaderXRequestID = "X-Request-Id"
)

func requestIDMiddleware(next HandlerFunc) HandlerFunc {
	return func(c Context) error {
		rid := c.Request().Header.Get(HeaderXRequestID)
		if rid == "" {
			rid = uuid.New().String()
			c.Request().Header.Set(HeaderXRequestID, rid)
		}
		c.Response().Header().Set(HeaderXRequestID, rid)
		return next(c)
	}
}

func addLoggerToRequestContext(log *zap.Logger) func(HandlerFunc) HandlerFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			ctx := logger.WithLogger(c.Request().Context(), log)
			request := c.Request().WithContext(ctx)
			c.SetRequest(request)
			return next(c)
		}
	}
}
