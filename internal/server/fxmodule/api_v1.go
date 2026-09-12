package fxmodule

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	"github.com/nikimonax/go-metrics/internal/server/handler"
)

func APIV1Module() fx.Option {
	return fx.Module(
		"api_v1",
		fx.Provide(
			fx.Annotate(
				handler.NewPreviewMetricsHandler,
				fx.ParamTags(
					"",
					`name:"presenter_error_text"`,
					`name:"presenter_metrics_html"`,
				),
				fx.ResultTags(`name:"handler_metrics_preview_v1"`),
			),
			fx.Annotate(
				handler.NewUpdateMetricHandler,
				fx.ParamTags("", `name:"presenter_error_text"`),
				fx.ResultTags(`name:"handler_metric_update_v1"`),
			),
			fx.Annotate(
				handler.NewGetMetricHandler,
				fx.ParamTags(
					"",
					`name:"presenter_error_text"`,
					`name:"presenter_metric_text"`,
				),
				fx.ResultTags(`name:"handler_metric_get_v1"`),
			),
			fx.Annotate(
				provideRouterV1,
				fx.ParamTags(
					`name:"router_base"`,
					`name:"middleware_logger"`,
					`name:"middleware_compress"`,
				),
				fx.ResultTags(`name:"router_v1"`),
			),
		),
		fx.Invoke(
			fx.Annotate(
				registerHandlersV1,
				fx.ParamTags(
					`name:"router_v1"`,
					`name:"handler_metrics_preview_v1"`,
					`name:"handler_metric_update_v1"`,
					`name:"handler_metric_get_v1"`,
				),
			),
		),
	)
}

func provideRouterV1(
	baseRouter chi.Router,
	logger httpextra.Middleware,
	compress httpextra.Middleware,
) chi.Router {
	return baseRouter.With(compress, logger)
}

func registerHandlersV1(
	router chi.Router,
	previewMetricsHandler http.Handler,
	updateMetricHandler http.Handler,
	getMetricHandler http.Handler,
) {
	router.Get(
		"/",
		previewMetricsHandler.ServeHTTP,
	)

	router.Post(
		"/update/{metricType}/{metricName}/{metricValue}",
		updateMetricHandler.ServeHTTP,
	)
	router.Get(
		"/value/{metricType}/{metricName}",
		getMetricHandler.ServeHTTP,
	)
}
