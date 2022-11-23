package limiter

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
	overlappedDuration := time.Since(p.end)
	if overlappedDuration >= p.duration {
		return 0
	}
	return uint64(float64(p.counters[string(ip)]) * float64(p.duration-overlappedDuration) / float64(p.duration))
}

func (p period) Get(ip net.IP) uint64 {
	return p.counters[string(ip)]
}

func (p period) Increment(ip net.IP) {
	p.counters[string(ip)]++
}

// NewWeightedWindowLimiter returns new limiter implementing weighted window algorithm
func NewWeightedWindowLimiter(limit uint64, duration time.Duration) *WeightedWindowLimiter {
	return &WeightedWindowLimiter{
		limit:    limit,
		duration: duration,
		current:  newPeriod(duration),
	}
}

// WeightedWindowLimiter imlements rate limiting using weighted window algorithm
type WeightedWindowLimiter struct {
	limit    uint64
	duration time.Duration

	mu       sync.Mutex
	previous period
	current  period
}

// IsRequestAllowed tells if request should be handled or rejected due to exhausted rate limit
func (l *WeightedWindowLimiter) IsRequestAllowed(ip net.IP) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	allowed := l.previous.GetProportionally(ip)+l.current.Get(ip) <= l.limit
	if allowed {
		l.current.Increment(ip)
	}
	return allowed
}

// Run runs cleaning task of the limiter
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
