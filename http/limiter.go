package http

import (
	nethttp "net/http"

	"github.com/pkg/errors"

	"github.com/CoreumFoundation/faucet/pkg/http"
	"github.com/CoreumFoundation/faucet/pkg/limiter"
)

func limiterMiddleware(limiter limiter.PerIPLimiter) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(c http.Context) error {
			r := c.Request()
			if r.Method == nethttp.MethodGet {
				return next(c)
			}

			ip, err := http.IPFromRequest(r)
			if err != nil {
				return err
			}
			if !ip.IsPrivate() && !ip.IsLoopback() && !ip.IsLinkLocalUnicast() && !limiter.IsRequestAllowed(ip) {
				return errors.Wrapf(ErrRateLimitExhausted, "ip %q has already used its rate limit", ip.String())
			}
			return next(c)
		}
	}
}
