package fxmodule

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	"github.com/nikimonax/go-metrics/internal/server/handler"
)

func APIDBModule() fx.Option {
	return fx.Module(
		"api_db",
		fx.Provide(
			fx.Annotate(
				handler.NewPingDatabaseHandler,
				fx.ResultTags(`name:"handler_database_ping"`),
				fx.As(new(http.Handler)),
			),
		),
		fx.Invoke(
			fx.Annotate(
				registerDatabasePingHandler,
				fx.ParamTags(
					`name:"router_base"`,
					`name:"middleware_logger"`,
					`name:"handler_database_ping"`,
				),
			),
		),
	)
}

func registerDatabasePingHandler(
	router chi.Router,
	logger httpextra.Middleware,
	pingHandler http.Handler,
) {
	router.With(logger).Get("/ping", pingHandler.ServeHTTP)
}
