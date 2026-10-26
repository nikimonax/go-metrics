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

// PreviewMetricsHandler godoc
//
//	@ID				previewMetrics
//	@Summary		Список всех метрик (HTML)
//	@Description	Возвращает HTML-страницу со списком всех метрик.
//	@Tags			metrics
//	@Produce		text/html
//	@Success		200	{string}	string	"HTML-страница со списком метрик"
//	@Failure		500	{object}	model.ErrorResponse
//	@Router			/ [get]
//
// ServeHTTP implements [http.Handler].
func (h *PreviewMetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.useCase.Execute(r.Context())

	if err != nil {
		h.errorPresenter.Render(w, err, http.StatusInternalServerError)
		return
	}

	h.metricsPresenter.Render(w, metrics, http.StatusOK)

}

var _ http.Handler = (*PreviewMetricsHandler)(nil)

func NewPreviewMetricsHandler(
	useCase GetAllMetricsUseCase,
	errorPresenter presenter.ErrorPresenter,
	metricsPresenter presenter.MetricsPresenter,
) *PreviewMetricsHandler {
	return &PreviewMetricsHandler{
		useCase:          useCase,
		errorPresenter:   errorPresenter,
		metricsPresenter: metricsPresenter,
	}
}
