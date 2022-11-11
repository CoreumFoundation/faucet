package http

import "github.com/pkg/errors"

// ErrRateLimitExhausted is returned when rate limit is exhausted for an IP address
var ErrRateLimitExhausted = errors.New("rate limit exhausted")
