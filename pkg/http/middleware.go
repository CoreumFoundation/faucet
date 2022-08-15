package http

import (
	"github.com/google/uuid"
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
