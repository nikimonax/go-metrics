package handler

import (
	"errors"
	"net/http"

	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/impl"
	"github.com/nikimonax/go-metrics/internal/server/presenter"
)

type GetMetricHandler struct {
	useCase         GetMetricUseCase
	errorPresenter  presenter.ErrorPresenter
	metricPresenter presenter.MetricPresenter
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
		status := http.StatusInternalServerError

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
	errorPresenter presenter.ErrorPresenter,
	metricPresenter presenter.MetricPresenter,
) http.Handler {
	return &GetMetricHandler{
		useCase:         useCase,
		errorPresenter:  errorPresenter,
		metricPresenter: metricPresenter,
	}
}
