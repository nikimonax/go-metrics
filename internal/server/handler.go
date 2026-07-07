package server

import (
	"errors"
	"net/http"

	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/impl"
	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
)

// update metric

type UpdateMetricUseCase interface {
	Execute(domain.Metric) error
}

type UpdateMetricHandler struct {
	useCase        UpdateMetricUseCase
	errorPresenter ErrorPresenter
}

// ServeHTTP implements [http.Handler].
func (h *UpdateMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metric, err := parseMetricFromRequest(r)

	if err != nil {
		h.errorPresenter.Render(w, err, http.StatusBadRequest)
		return
	}

	err = h.useCase.Execute(metric)

	if err != nil {
		h.errorPresenter.Render(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set(httpextra.HDRContentType, httpextra.MIMEText)
	w.WriteHeader(http.StatusOK)
}

func NewUpdateMetricHandler(
	useCase UpdateMetricUseCase,
	errorPresenter ErrorPresenter,
) http.Handler {
	return &UpdateMetricHandler{
		useCase:        useCase,
		errorPresenter: errorPresenter,
	}
}

// get metric

type GetMetricUseCase interface {
	Execute(
		metricType domain.MetricType,
		metricName domain.MetricName,
	) (domain.Metric, error)
}

type GetMetricHandler struct {
	useCase         GetMetricUseCase
	errorPresenter  ErrorPresenter
	metricPresenter MetricPresenter
}

func (h *GetMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		metricType domain.MetricType
		metricName domain.MetricName
	)

	if err := parseMetricType(r, &metricType); err != nil {
		h.errorPresenter.Render(w, err, http.StatusBadRequest)
		return
	}

	if err := parseMetricName(r, &metricName); err != nil {
		h.errorPresenter.Render(w, err, http.StatusBadRequest)
		return
	}

	metric, err := h.useCase.Execute(metricType, metricName)

	if err != nil {
		status := http.StatusServiceUnavailable

		if errors.Is(err, impl.ErrMetricNotFound) {
			status = http.StatusNotFound
		}

		h.errorPresenter.Render(w, err, status)
		return

	}

	h.metricPresenter.Render(w, metric, http.StatusOK)
}

func NewGetMetricHandler(
	useCase GetMetricUseCase,
	errorPresenter ErrorPresenter,
	metricPresenter MetricPresenter,
) http.Handler {
	return &GetMetricHandler{
		useCase:         useCase,
		errorPresenter:  errorPresenter,
		metricPresenter: metricPresenter,
	}
}

// preview metrics

type GetAllMetricsUseCase interface {
	Execute() ([]domain.Metric, error)
}

type PreviewMetricsHandler struct {
	useCase          GetAllMetricsUseCase
	errorPresenter   ErrorPresenter
	metricsPresenter MetricsPresenter
}

func (h *PreviewMetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.useCase.Execute()

	if err != nil {
		h.errorPresenter.Render(w, err, http.StatusInternalServerError)
		return
	}

	h.metricsPresenter.Render(w, metrics, http.StatusOK)

}

func NewPreviewMetricsHandler(
	useCase GetAllMetricsUseCase,
	errorPresenter ErrorPresenter,
	metricsPresenter MetricsPresenter,
) http.Handler {
	return &PreviewMetricsHandler{
		useCase:          useCase,
		errorPresenter:   errorPresenter,
		metricsPresenter: metricsPresenter,
	}
}
