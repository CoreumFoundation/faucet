package http

import "github.com/pkg/errors"

var ErrRateLimitExhausted = errors.New("rate limit exhausted")
