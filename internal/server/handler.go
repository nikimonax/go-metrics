package server

import (
	"net/http"

	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
)

type UpdateMetricUseCase interface {
	Execute(domain.Metric) error
}

type UpdateMetricHandler struct {
	useCase UpdateMetricUseCase
}

// ServeHTTP implements [http.Handler].
func (h *UpdateMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get(httpextra.HDRContentType)

	if contentType != httpextra.MIMEText {
		message := newErrMsgInvalidContentType(contentType, httpextra.MIMEText)
		http.Error(w, message, http.StatusBadRequest)
		return
	}

	metric, err := parseMetricFromRequest(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.useCase.Execute(metric)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(httpextra.HDRContentType, httpextra.MIMEText)
	w.WriteHeader(http.StatusOK)
}

func NewUpdateMetricHandler(useCase UpdateMetricUseCase) http.Handler {
	return &UpdateMetricHandler{useCase: useCase}
}

// get metric

type GetMetricUseCase interface {
	Execute(
		metricType domain.MetricType,
		metricName domain.MetricName,
	) (domain.Metric, error)
}

type GetMetricHandler struct {
	useCase GetMetricUseCase
}

func (h *GetMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		metricType domain.MetricType
		metricName domain.MetricName
	)

	if err := parseMetricType(r, &metricType); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := parseMetricName(r, &metricName); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metric, err := h.useCase.Execute(metricType, metricName)

	if err != nil {
		status := http.StatusServiceUnavailable

		if errors.Is(err, impl.ErrMetricNotFound) {
			status = http.StatusNotFound
		}

		http.Error(w, err.Error(), status)
		return

	}

	w.Header().Set(httpextra.HDRContentType, httpextra.MIMEText)
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte(metric.Value().String())); err != nil {
		log.Printf("Failed to write HTTP response: %v", err)
	}
}

func NewGetMetricHandler(useCase GetMetricUseCase) http.Handler {
	return &GetMetricHandler{useCase: useCase}
}
