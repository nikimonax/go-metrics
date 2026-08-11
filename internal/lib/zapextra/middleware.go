package zapextra

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func NewZapSugarLoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	sugar := logger.Sugar()

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			t1 := time.Now()

			defer func() {
				sugar.Infow(
					"handle request",
					"method", r.Method,
					"uri", r.RequestURI,
					"status", ww.Status(),
					"duration", time.Since(t1),
					"size", ww.BytesWritten(),
				)
			}()

			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}
