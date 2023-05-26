package limiter

import "net"

// PerIPLimiter defines an interface of IP rate limiter.
type PerIPLimiter interface {
	IsRequestAllowed(ip net.IP) bool
	Increment(ip net.IP)
}
