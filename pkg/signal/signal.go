package signal

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// TerminateSignal returns a signal
func TerminateSignal(opts ...opt) context.Context {
	options := defaultOptions
	for _, o := range opts {
		options = o(options)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-sigChan
		cancel()
		if options.ForceTimeout > 0 {
			go func() {
				time.Sleep(options.ForceTimeout)
				os.Exit(0)
			}()
		}
		if options.ForceOnSecondSignal {
			go func() {
				<-sigChan
				os.Exit(0)
			}()
		}
	}()
	return ctx
}

var defaultOptions = Options{
	ForceTimeout:        120 * time.Second,
	ForceOnSecondSignal: true,
}

// Options includes options for signal processing
type Options struct {
	ForceTimeout        time.Duration
	ForceOnSecondSignal bool
}

type opt = func(Options) Options

// WithForceTimeout instructs the application for abandon graceful shutdown
// and forcefully exit after a timeout if shutdown has not been successful.
func WithForceTimeout(duration time.Duration) opt {
	return func(o Options) Options {
		o.ForceTimeout = duration
		return o
	}
}

// WithForceOnSecondSignal instructs the application for abandon graceful shutdown
// and forcefully exit
func WithForceOnSecondSignal(force bool) opt {
	return func(o Options) Options {
		o.ForceOnSecondSignal = force
		return o
	}
}
