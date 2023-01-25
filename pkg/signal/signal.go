package signal

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
)

// TerminateSignal returns a context which will be cancelled if SIGINT or SIGTERM is received by the application.
func TerminateSignal(ctx context.Context) context.Context {
	log := logger.Get(ctx)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		s := <-sigChan
		log.Info("Received signal", zap.Stringer("signal", s))
		cancel()
	}()
	return ctx
}
