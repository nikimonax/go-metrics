package handler

import (
	"net/http"

	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	"github.com/nikimonax/go-metrics/internal/server/presenter"
)

type UpdateMetricHandler struct {
	useCase        UpdateMetricUseCase
	errorPresenter presenter.ErrorPresenter
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
	errorPresenter presenter.ErrorPresenter,
) http.Handler {
	return &UpdateMetricHandler{
		useCase:        useCase,
		errorPresenter: errorPresenter,
	}
}
