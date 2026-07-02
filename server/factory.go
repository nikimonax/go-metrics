package server

import (
	"net/http"

	"github.com/nikimonax/go-metrics/server/app"
	metricHttp "github.com/nikimonax/go-metrics/server/http"
	"github.com/nikimonax/go-metrics/server/internal"
)

func New() http.Handler {
	metricRepository := internal.NewInMemoryMetricRepository()

	updateMetricUseCase := app.NewUpdateMetricUseCase(metricRepository)
	updateMetricHandler := metricHttp.NewUpdateMetricHandler(updateMetricUseCase)

	mux := http.NewServeMux()
	mux.Handle(
		"/update/{metricType}/{metricName}/{metricValue}",
		updateMetricHandler,
	)

	return mux
}
