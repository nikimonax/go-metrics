package zapextra

import (
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewFxLogger(logger *zap.Logger) fxevent.Logger {
	fxLogger := &fxevent.ZapLogger{Logger: logger}
	fxLogger.UseLogLevel(zapcore.DebugLevel)
	return fxLogger
}
