package server

import (
	"context"

	"go.uber.org/fx"

	"github.com/nikimonax/go-metrics/internal/lib/zapextra"
	"github.com/nikimonax/go-metrics/internal/server/config"
	"github.com/nikimonax/go-metrics/internal/server/fxmodule"
	"github.com/nikimonax/go-metrics/internal/shared/fxutil"
)

type Server struct {
	app *fx.App
}

func (server *Server) Run(ctx context.Context) error {
	return fxutil.RunWithContext(ctx, server.app)
}

func New(cfg *config.ServerConfig) (*Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	app := fx.New(
		fx.Supply(cfg),
		fxmodule.StorageModule(cfg),
		fxmodule.CoreModule(),
		fxmodule.ProfileModule(cfg),
		fxmodule.APIV1Module(),
		fxmodule.APIV2Module(),
		fxmodule.APIV3Module(),
		fxmodule.SwaggerModule(),
		fxmodule.DumpModule(),
		fxmodule.AuditModule(cfg),
		fx.WithLogger(zapextra.NewFxLogger),
	)

	if err := app.Err(); err != nil {
		return nil, err
	}

	return &Server{app: app}, nil
}
