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

func BuildOpts(cfg *config.ServerConfig) []fx.Option {
	// base modules
	opts := []fx.Option{
		fx.Supply(cfg),
		fxmodule.CoreModule(),
		fxmodule.StorageModule(cfg),
		fxmodule.DumpModule(),
		fxmodule.AuditModule(cfg),
	}

	switch cfg.ServerType {
	case config.HTTPServer:
		// http specific features

		opts = append(
			opts,
			fxmodule.HTTPModule(),
		)

		if cfg.Profile {
			opts = append(opts, fxmodule.ProfileModule())
		}

		if cfg.Database.DSN != "" {
			opts = append(opts, fxmodule.APIDBModule())
		}

		opts = append(
			opts,
			fxmodule.SwaggerModule(),
			fxmodule.APIV1Module(),
			fxmodule.APIV2Module(),
			fxmodule.APIV3Module(),
		)

	case config.GRPCServer:
		// grpc specific features
		opts = append(
			opts,
			fxmodule.GrpcModule(),
		)
	}

	// final modules
	opts = append(
		opts,
		fxmodule.ServerModule(),
		fx.WithLogger(zapextra.NewFxLogger),
	)

	return opts
}

func New(cfg *config.ServerConfig) (*Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	app := fx.New(BuildOpts(cfg)...)

	if err := app.Err(); err != nil {
		return nil, err
	}

	return &Server{app: app}, nil
}
