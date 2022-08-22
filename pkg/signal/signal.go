package signal

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"go.uber.org/zap"
)

// TerminateSignal returns a context which will be cancelled if SIGINT or SIGTERM is received by the application
func TerminateSignal(ctx context.Context) context.Context {
	log := logger.Get(ctx)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		s := <-sigChan
		log.Info("received syscall", zap.Stringer("syscall", s))
		cancel()
	}()
	return ctx
}
