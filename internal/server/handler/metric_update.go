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

// UpdateMetricHandler godoc
//
//	@ID				updateMetricV1
//	@Summary		Обновление метрики (URL-параметры)
//	@Description	Обновляет значение метрики, переданной в URL-параметрах.
//	@Tags			metrics
//	@Accept			plain
//	@Produce		text/plain
//	@Param			metricType	path		string	true	"Тип метрики: gauge | counter"	Enums(gauge, counter)
//	@Param			metricName	path		string	true	"Имя метрики"
//	@Param			metricValue	path		number	true	"Значение метрики"
//	@Success		200			{string}	string	"OK"
//	@Failure		400			{object}	model.ErrorResponse
//	@Failure		500			{object}	model.ErrorResponse
//	@Router			/update/{metricType}/{metricName}/{metricValue} [post]
//
// ServeHTTP implements [http.Handler].
func (h *UpdateMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metric, err := parseMetricFromRequest(r)

	if err != nil {
		h.errorPresenter.Render(w, err, http.StatusBadRequest)
		return
	}

	err = h.useCase.Execute(r.Context(), metric)

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
