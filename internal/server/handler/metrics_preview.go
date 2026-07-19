package handler

import (
	"net/http"

	"github.com/nikimonax/go-metrics/internal/server/presenter"
)

type PreviewMetricsHandler struct {
	useCase          GetAllMetricsUseCase
	errorPresenter   presenter.ErrorPresenter
	metricsPresenter presenter.MetricsPresenter
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
	errorPresenter presenter.ErrorPresenter,
	metricsPresenter presenter.MetricsPresenter,
) http.Handler {
	return &PreviewMetricsHandler{
		useCase:          useCase,
		errorPresenter:   errorPresenter,
		metricsPresenter: metricsPresenter,
	}
}
