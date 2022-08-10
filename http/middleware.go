package http

import (
	"net/http"
	"time"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"go.uber.org/zap"
)

type middleware struct {
	next http.Handler
	log  *zap.Logger
}

func (m middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	log := m.log.With(
		zap.String("method", r.Method),
		zap.String("x_forwarded_for", r.Header.Get("x-forwarded-for")),
		zap.String("path", r.URL.Path),
	)
	ctx := logger.WithLogger(r.Context(), log)
	r = r.WithContext(ctx)
	defer func() {
		duration := time.Since(started)
		log.Debug("request served",
			zap.Duration("duration", duration),
		)
	}()
	m.next.ServeHTTP(w, r)
}
