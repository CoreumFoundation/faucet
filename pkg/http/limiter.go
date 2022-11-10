package http

// This implements one of the variants of limiting request rate described at
// https://blog.cloudflare.com/counting-things-a-lot-of-different-things/

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/pkg/errors"
)

func newPeriod(duration time.Duration) period {
	return period{
		duration: duration,
		end:      time.Now().Add(duration),
		counters: map[string]uint64{},
	}
}

type period struct {
	duration time.Duration
	end      time.Time
	counters map[string]uint64
}

func (p period) GetProportionally(ip net.IP) uint64 {
	if p.duration.Nanoseconds() == 0 {
		return 0
	}
	overlappedDuration := time.Now().Sub(p.end)
	if overlappedDuration >= p.duration {
		return 0
	}
	return uint64(float64(p.counters[string(ip)]) * float64(p.duration-overlappedDuration) / float64(p.duration))
}

func (p period) Increment(ip net.IP) uint64 {
	p.counters[string(ip)]++
	return p.counters[string(ip)]
}

func NewWeightedWindowLimiter(limit uint64, duration time.Duration) *WeightedWindowLimiter {
	return &WeightedWindowLimiter{
		limit:    limit,
		duration: duration,
		current:  newPeriod(duration),
	}
}

type WeightedWindowLimiter struct {
	limit    uint64
	duration time.Duration

	mu       sync.Mutex
	previous period
	current  period
}

func (l *WeightedWindowLimiter) IsRequestAllowed(ip net.IP) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.previous.GetProportionally(ip)+l.current.Increment(ip) <= l.limit
}

func (l *WeightedWindowLimiter) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return errors.WithStack(ctx.Err())
		case <-time.After(l.duration):
			l.mu.Lock()
			l.previous = l.current
			l.current = newPeriod(l.duration)
			l.mu.Unlock()
		}
	}
}
