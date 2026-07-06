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
