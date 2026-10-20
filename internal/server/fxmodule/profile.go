package fxmodule

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	"github.com/nikimonax/go-metrics/internal/server/config"
	"github.com/nikimonax/go-metrics/internal/server/handler"
)

func ProfileModule(cfg *config.ServerConfig) fx.Option {
	if !cfg.Profile {
		return fx.Options()
	}

	return fx.Module(
		"profile",
		fx.Invoke(
			fx.Annotate(
				registerProfiler,
				fx.ParamTags(
					`name:"router_base"`,
					`name:"middleware_logger"`,
				),
			),
		),
	)
}

func registerProfiler(
	baseRouter chi.Router,
	logger httpextra.Middleware,
) {
	baseRouter.With(logger).Mount("/debug", handler.Profiler())
}
