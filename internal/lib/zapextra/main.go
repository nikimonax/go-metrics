package zapextra

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type LogEnv = int

const (
	EnvDev LogEnv = iota
	EnvProd
)

func NewZapLogger(env LogEnv) *zap.Logger {
	var config zap.Config

	switch env {
	case EnvDev:
		config = zap.NewDevelopmentConfig()
	case EnvProd:
		config = zap.NewProductionConfig()
	default:
		log.Fatalf("unknown env")
	}

	config.DisableCaller = true

	logger, err := config.Build()

	if err != nil {
		log.Fatalf("failed create zap logger: %s", err)
	}

	return logger
}

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
