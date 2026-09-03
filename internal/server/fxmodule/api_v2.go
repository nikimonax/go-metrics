package fxmodule

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	"github.com/nikimonax/go-metrics/internal/server/handler"
)

func APIV2Module() fx.Option {
	return fx.Module(
		"api_v2",
		fx.Provide(
			fx.Annotate(
				handler.NewUpdateMetricV2Handler,
				fx.ParamTags("", `name:"presenter_error_json"`, ""),
				fx.ResultTags(`name:"handler_metric_update_v2"`),
			),
			fx.Annotate(
				handler.NewGetMetricV2Handler,
				fx.ParamTags(
					"",
					`name:"presenter_error_json"`,
					`name:"presenter_metric_json"`,
					"",
				),
				fx.ResultTags(`name:"handler_metric_get_v2"`),
			),
			fx.Annotate(
				provideJSONRouter,
				fx.ParamTags(
					`name:"router_base"`,
					`name:"middleware_logger"`,
					`name:"middleware_compress"`,
				),
				fx.ResultTags(`name:"router_v2"`),
			),
		),
		fx.Invoke(
			fx.Annotate(
				registerHandlersV2,
				fx.ParamTags(
					`name:"router_v2"`,
					`name:"handler_metric_update_v2"`,
					`name:"handler_metric_get_v2"`,
				),
			),
		),
	)
}

func registerHandlersV2(
	router chi.Router,
	updateMetricHandler http.Handler,
	getMetricHandler http.Handler,
) {
	router.Post(
		"/update",
		updateMetricHandler.ServeHTTP,
	)
	router.Post(
		"/value",
		getMetricHandler.ServeHTTP,
	)

}
