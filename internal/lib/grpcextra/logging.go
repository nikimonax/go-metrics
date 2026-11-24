package grpcextra

import (
	"context"
	"fmt"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func mapLevel(lvl logging.Level) zapcore.Level {
	switch lvl {
	case logging.LevelDebug:
		return zapcore.DebugLevel
	case logging.LevelInfo:
		return zapcore.InfoLevel
	case logging.LevelWarn:
		return zapcore.WarnLevel
	case logging.LevelError:
		return zapcore.ErrorLevel
	default:
		panic(fmt.Sprintf("unknown level %v", lvl))
	}
}

func InterceptorSugarLogger(sugar *zap.SugaredLogger) logging.Logger {
	fn := func(_ context.Context, lvl logging.Level, msg string, fields ...any) {
		sugar.Logw(mapLevel(lvl), msg, fields...)
	}
	return logging.LoggerFunc(fn)
}
