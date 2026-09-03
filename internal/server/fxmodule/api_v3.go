package fxmodule

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	"github.com/nikimonax/go-metrics/internal/server/handler"
)

func APIV3Module() fx.Option {
	return fx.Module(
		"api_v3",
		fx.Provide(
			fx.Annotate(
				handler.NewUpdateMetricsV3Handler,
				fx.ParamTags("", `name:"presenter_error_json"`, ""),
				fx.ResultTags(`name:"handler_metrics_update_v3"`),
			),
			fx.Annotate(
				provideJSONRouter,
				fx.ParamTags(
					`name:"router_base"`,
					`name:"middleware_logger"`,
					`name:"middleware_compress"`,
				),
				fx.ResultTags(`name:"router_v3"`),
			),
		),
		fx.Invoke(
			fx.Annotate(
				registerHandlersV3,
				fx.ParamTags(
					`name:"router_v3"`,
					`name:"handler_metrics_update_v3"`,
				),
			),
		),
	)
}

func registerHandlersV3(
	router chi.Router,
	updateMetricsHandler http.Handler,
) {
	router.Post(
		"/updates",
		updateMetricsHandler.ServeHTTP,
	)
}
