package http

import (
	"net/http"

	"github.com/nikimonax/go-metrics/server/app"
)

type UpdateMetricHandler struct {
	metricService *app.MetricService
}

// ServeHTTP implements [http.Handler].
func (h *UpdateMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		message := newErrMsgInvalidMethod(r.Method, http.MethodPost)
		http.Error(w, message, http.StatusMethodNotAllowed)
		return
	}

	if contentType := r.Header.Get("Content-Type"); contentType != "text/plain" {
		message := newErrMsgInvalidContentType(contentType, "text/plain")
		http.Error(w, message, http.StatusBadRequest)
		return
	}

	var dto app.UpdateMetricDTO

	if err := parseUpdateMetricDTO(r, &dto); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.metricService.Update(dto); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func NewUpdateMetricHandler(metricService *app.MetricService) http.Handler {
	return &UpdateMetricHandler{
		metricService: metricService,
	}
}
