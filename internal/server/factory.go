package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nikimonax/go-metrics/internal/app"
	"github.com/nikimonax/go-metrics/internal/impl"
)

func New() http.Handler {
	metricRepository := impl.NewInMemoryMetricRepository()

	updateMetricUseCase := app.NewUpdateMetricUseCase(metricRepository)
	getMetricUseCase := app.NewGetMetricUseCase(metricRepository)
	// getAllMetricsUseCase := app.NewGetAllMetricsUseCase(metricRepository)

	plainTextErrorPresenter := NewPlainTextErrorPresenter()
	plainTextMetricPresenter := NewPlainTextMetricPresenter()
	// htmlTableMetricsPresenter := NewHtmlTableMetricsPresenter()

	updateMetricHandler := NewUpdateMetricHandler(
		updateMetricUseCase,
		plainTextErrorPresenter,
	)
	getMetricHandler := NewGetMetricHandler(
		getMetricUseCase,
		plainTextErrorPresenter,
		plainTextMetricPresenter,
	)
	// PreviewMetricsHandler := NewPreviewMetricsHandler(
	// 	getAllMetricsUseCase,
	// 	plainTextErrorPresenter,
	// 	htmlTableMetricsPresenter,
	// )

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// r.Get("/", PreviewMetricsHandler.ServeHTTP)
	r.Post(
		"/update/{metricType}/{metricName}/{metricValue}",
		updateMetricHandler.ServeHTTP,
	)
	r.Get(
		"/value/{metricType}/{metricName}",
		getMetricHandler.ServeHTTP,
	)

	return r
}
