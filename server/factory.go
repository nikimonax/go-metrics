package server

import (
	"net/http"

	"github.com/nikimonax/go-metrics/server/app"
	metricHttp "github.com/nikimonax/go-metrics/server/http"
	"github.com/nikimonax/go-metrics/server/internal"
)

func New() http.Handler {
	metricRepository := internal.NewInMemoryMetricRepository()
	metricService := app.NewMetricService(metricRepository)

	updateMetricHandler := metricHttp.NewUpdateMetricHandler(metricService)

	mux := http.NewServeMux()
	mux.Handle(
		"/update/{metricType}/{metricName}/{metricValue}",
		updateMetricHandler,
	)

	return mux
}
