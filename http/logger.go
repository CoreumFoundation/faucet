package http

import (
	"net/http"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"go.uber.org/zap"
)

type loggerMiddleware struct {
	next http.Handler
	log  *zap.Logger
}

func (m loggerMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := logger.WithLogger(r.Context(), m.log)
	r = r.WithContext(ctx)
	m.next.ServeHTTP(w, r)
}
