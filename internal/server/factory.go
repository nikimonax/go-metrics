package server

import (
	"go.uber.org/fx"

	"github.com/nikimonax/go-metrics/internal/lib/zapextra"
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
		fxmodule.StorageModule(cfg),
		fxmodule.CoreModule(),
		fxmodule.APIV1Module(),
		fxmodule.APIV2Module(),
		fxmodule.APIV3Module(),
		fxmodule.DumpModule(),
		fx.WithLogger(zapextra.NewFxLogger),
	)

	if err := app.Err(); err != nil {
		return nil, err
	}

	return &Server{app: app}, nil
}
