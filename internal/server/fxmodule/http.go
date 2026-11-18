package fxmodule

import (
	"compress/gzip"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	"github.com/nikimonax/go-metrics/internal/lib/compress"
	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	"github.com/nikimonax/go-metrics/internal/lib/httpsec"
	"github.com/nikimonax/go-metrics/internal/lib/zapextra"
	"github.com/nikimonax/go-metrics/internal/server/config"
	"github.com/nikimonax/go-metrics/internal/server/middleware"
	"github.com/nikimonax/go-metrics/internal/server/presenter"
)

func HTTPModule() fx.Option {
	return fx.Module(
		"http",
		// presenters
		fx.Provide(
			fx.Annotate(
				presenter.NewPlainTextErrorPresenter,
				fx.ResultTags(`name:"presenter_error_text"`),
				fx.As(new(presenter.ErrorPresenter)),
			),
			fx.Annotate(
				presenter.NewPlainTextMetricPresenter,
				fx.ResultTags(`name:"presenter_metric_text"`),
				fx.As(new(presenter.MetricPresenter)),
			),
			fx.Annotate(
				presenter.NewHTMLTableMetricsPresenter,
				fx.ResultTags(`name:"presenter_metrics_html"`),
				fx.As(new(presenter.MetricsPresenter)),
			),
			fx.Annotate(
				presenter.NewJSONErrorPresenter,
				fx.ResultTags(`name:"presenter_error_json"`),
				fx.As(new(presenter.ErrorPresenter)),
			),
			fx.Annotate(
				presenter.NewJSONMetricPresenter,
				fx.ResultTags(`name:"presenter_metric_json"`),
				fx.As(new(presenter.MetricPresenter)),
			),
		),
		// http routes, handlers, middleware
		fx.Provide(
			fx.Annotate(
				zapextra.NewZapSugarLoggingMiddleware,
				fx.ResultTags(`name:"middleware_logger"`),
			),
			fx.Annotate(
				provideMiddlewareCompress,
				fx.ResultTags(`name:"middleware_compress"`),
			),
			fx.Annotate(
				provideBaseRouter,
				fx.ResultTags(`name:"router_base"`),
			),
			fx.Annotate(
				provideHTTPServer,
				fx.ParamTags("", `name:"router_base"`),
			),
		),
	)
}

func provideMiddlewareCompress() httpextra.Middleware {
	compressor := middleware.NewCompressor(gzip.BestSpeed)
	compressor.SetEncoder("deflate", compress.EncoderDeflate)
	compressor.SetEncoder("gzip", compress.EncoderGzip)
	return compressor.Handler
}

func provideBaseRouter(
	cfg *config.ServerConfig,
	hasher *httpsec.Hasher,
) (chi.Router, error) {
	baseRouter := chi.NewRouter()

	baseRouter.Use(
		middleware.CleanPath,
		middleware.ClientIPFromHeader(config.SourceIPHeader),
	)

	if cfg.TrustedSubnet != "" {
		allowIPsMiddleware, err := middleware.AllowIPs(cfg.TrustedSubnet)

		if err != nil {
			return nil, err
		}

		baseRouter.Use(allowIPsMiddleware)
	}

	if hasher != nil {
		baseRouter.Use(hasher.CalculateMiddleware())
	}

	return baseRouter, nil
}

func provideHTTPServer(
	cfg *config.ServerConfig,
	router chi.Router,
) server {
	return &http.Server{
		Addr:    cfg.Listen,
		Handler: router,
	}
}
