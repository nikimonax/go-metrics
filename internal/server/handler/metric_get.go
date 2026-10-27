package handler

import (
	"errors"
	"net/http"

	"github.com/nikimonax/go-metrics/internal/app"
	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/server/presenter"
)

type GetMetricHandler struct {
	useCase         GetMetricUseCase
	errorPresenter  presenter.ErrorPresenter
	metricPresenter presenter.MetricPresenter
}

// GetMetricHandler godoc
//
//	@ID				getMetricV1
//	@Summary		Получение значения метрики (URL-параметры)
//	@Description	Возвращает значение метрики в текстовом виде.
//	@Tags			metrics
//	@Produce		text/plain
//	@Param			metricType	path		string	true	"Тип метрики: gauge | counter"	Enums(gauge, counter)
//	@Param			metricName	path		string	true	"Имя метрики"
//	@Success		200			{string}	string	"Значение метрики"
//	@Failure		400			{object}	model.ErrorResponse
//	@Failure		404			{object}	model.ErrorResponse
//	@Failure		500			{object}	model.ErrorResponse
//	@Router			/value/{metricType}/{metricName} [get]
//
// ServeHTTP implements [http.Handler].
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

	metric, err := h.useCase.Execute(
		r.Context(),
		metricType,
		metricName,
	)

	if err != nil {
		status := http.StatusInternalServerError

		if errors.Is(err, app.ErrMetricNotFound) {
			status = http.StatusNotFound
		}

		h.errorPresenter.Render(w, err, status)
		return

	}

	h.metricPresenter.Render(w, metric, http.StatusOK)
}

var _ http.Handler = (*GetMetricHandler)(nil)

func NewGetMetricHandler(
	useCase GetMetricUseCase,
	errorPresenter presenter.ErrorPresenter,
	metricPresenter presenter.MetricPresenter,
) *GetMetricHandler {
	return &GetMetricHandler{
		useCase:         useCase,
		errorPresenter:  errorPresenter,
		metricPresenter: metricPresenter,
	}
}
