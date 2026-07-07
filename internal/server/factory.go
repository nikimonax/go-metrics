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
	updateMetricHandler := NewUpdateMetricHandler(updateMetricUseCase)

	getMetricUseCase := app.NewGetMetricUseCase(metricRepository)
	getMetricHandler := NewGetMetricHandler(getMetricUseCase)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

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
