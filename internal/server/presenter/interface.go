package presenter

import (
	"net/http"

	"github.com/nikimonax/go-metrics/internal/domain"
)

type ErrorPresenter interface {
	Render(w http.ResponseWriter, err error, code int)
}

type MetricPresenter interface {
	Render(w http.ResponseWriter, metric domain.Metric, code int)
}

type MetricsPresenter interface {
	Render(w http.ResponseWriter, metrics []domain.Metric, code int)
}
