package signal

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// TerminateSignal returns a context which will be cancelled if SIGINT or SIGTERM is received by the application
func TerminateSignal(ctx context.Context) context.Context {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		<-sigChan
		cancel()
	}()
	return ctx
}
