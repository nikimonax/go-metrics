package server

import (
	"net/http"

	"github.com/nikimonax/go-metrics/internal/app"
	"github.com/nikimonax/go-metrics/internal/impl"
)

func New() http.Handler {
	metricRepository := impl.NewInMemoryMetricRepository()

	updateMetricUseCase := app.NewUpdateMetricUseCase(metricRepository)
	updateMetricHandler := NewUpdateMetricHandler(updateMetricUseCase)

	mux := http.NewServeMux()
	mux.Handle(
		"/update/{metricType}/{metricName}/{metricValue}",
		updateMetricHandler,
	)

	return mux
}
