package fxmodule

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	"github.com/nikimonax/go-metrics/internal/server/handler"
)

func SwaggerModule() fx.Option {
	return fx.Module(
		"docs",
		fx.Invoke(
			fx.Annotate(
				registerSwaggerUI,
				fx.ParamTags(`name:"router_base"`),
			),
		),
	)
}

func registerSwaggerUI(router chi.Router) {
	router.Mount("/swagger", handler.SwaggerHandler)
}
