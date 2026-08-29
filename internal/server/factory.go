package server

import (
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/nikimonax/go-metrics/internal/server/config"
	"github.com/nikimonax/go-metrics/internal/server/fxmodule"
)

type Server struct {
	app *fx.App
}

func (server *Server) Run() {
	server.app.Run()
}

func New(cfg *config.ServerConfig) (*Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	app := fx.New(
		fx.Supply(cfg),
		fxmodule.CoreModule(),
		fxmodule.DatabaseModule(cfg),
		fxmodule.APIV1Module(),
		fxmodule.APIV2Module(),
		fxmodule.DumpModule(),
		fx.WithLogger(provideFxLogger),
	)

	if err := app.Err(); err != nil {
		return nil, err
	}

	return &Server{app: app}, nil
}

func provideFxLogger(logger *zap.Logger) fxevent.Logger {
	fxLogger := &fxevent.ZapLogger{Logger: logger}
	fxLogger.UseLogLevel(zapcore.DebugLevel)
	return fxLogger
}
