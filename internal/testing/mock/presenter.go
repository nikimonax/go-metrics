package mock

import (
	"net/http"

	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/server/presenter"
	"github.com/stretchr/testify/mock"
)

// error presenter

type ErrorPresenter struct {
	mock.Mock
}

// Render implements [server.ErrorPresenter].
func (presenter *ErrorPresenter) Render(w http.ResponseWriter, err error, code int) {
	presenter.Called(w, err, code)
}

var _ presenter.ErrorPresenter = (*ErrorPresenter)(nil)

// metric presenter

type MetricPresenter struct {
	mock.Mock
}

// Render implements [server.MetricPresenter].
func (presenter *MetricPresenter) Render(w http.ResponseWriter, metric domain.Metric, code int) {
	presenter.Called(w, metric, code)
}

var _ presenter.MetricPresenter = (*MetricPresenter)(nil)

// metrics presenter

type MetricsPresenter struct {
	mock.Mock
}

// Render implements [server.MetricsPresenter].
func (presenter *MetricsPresenter) Render(w http.ResponseWriter, metrics []domain.Metric, code int) {
	presenter.Called(w, metrics, code)
}

var _ presenter.MetricsPresenter = (*MetricsPresenter)(nil)
