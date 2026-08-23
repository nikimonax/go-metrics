package fxmodule

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"

	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	"github.com/nikimonax/go-metrics/internal/server/handler"
	mymiddleware "github.com/nikimonax/go-metrics/internal/server/middleware"
	"github.com/nikimonax/go-metrics/internal/server/presenter"
)

func APIV2Module() fx.Option {
	return fx.Module(
		"api_v2",
		fx.Provide(
			provideValidator,
			provideTranslator,
		),
		fx.Provide(
			fx.Annotate(
				presenter.NewJSONErrorPresenter,
				fx.ResultTags(`name:"presenter_error_json"`),
			),
			fx.Annotate(
				presenter.NewJSONMetricPresenter,
				fx.ResultTags(`name:"presenter_metric_json"`),
			),
		),
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
				provideRouterV2,
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
					"",
					`name:"handler_metric_update_v2"`,
					`name:"handler_metric_get_v2"`,
				),
			),
		),
	)
}

func provideValidator() *validator.Validate {
	return validator.New(validator.WithRequiredStructEnabled())
}

func provideTranslator() ut.Translator {
	return presenter.NewTranslator()
}

func provideRouterV2(
	baseRouter chi.Router,
	logger Middleware,
	compress Middleware,
) chi.Router {
	return baseRouter.With(
		middleware.AllowContentType(httpextra.MIMEJSON),
		middleware.AllowContentEncoding(httpextra.ENCGzip),
		mymiddleware.Decompress(),
		compress,
		logger,
	)
}

func registerHandlersV2(
	router chi.Router,
	updateMetricHook *mymiddleware.RequestHook,
	updateMetricHandler http.Handler,
	getMetricHandler http.Handler,
) {
	router.With(
		updateMetricHook.Middleware,
	).Post(
		"/update",
		updateMetricHandler.ServeHTTP,
	)
	router.Post(
		"/value",
		getMetricHandler.ServeHTTP,
	)

}
