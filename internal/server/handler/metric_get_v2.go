package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/nikimonax/go-metrics/internal/app"
	"github.com/nikimonax/go-metrics/internal/model"
	"github.com/nikimonax/go-metrics/internal/server/presenter"
)

type GetMetricV2Handler struct {
	useCase         GetMetricUseCase
	errorPresenter  presenter.ErrorPresenter
	metricPresenter presenter.MetricPresenter
	validate        *validator.Validate
}

// GetMetricV2Handler godoc
//
//	@ID				getMetricV2
//	@Summary		Получение значения метрики (JSON)
//	@Description	Возвращает метрику в формате JSON.
//	@Tags			metrics
//	@Accept			json
//	@Produce		json
//	@Param			metricInfo	body		model.MetricInfo	true	"Идентификатор метрики"
//	@Success		200			{object}	model.Metric
//	@Failure		400			{object}	model.ErrorResponse
//	@Failure		404			{object}	model.ErrorResponse
//	@Failure		500			{object}	model.ErrorResponse
//	@Router			/value [post]
//
// ServeHTTP implements [http.Handler].
func (h *GetMetricV2Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var payload model.MetricInfo

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&payload); err != nil {
		h.errorPresenter.Render(w, err, http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(&payload); err != nil {
		httpStatus := http.StatusInternalServerError

		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			httpStatus = http.StatusBadRequest
		}

		h.errorPresenter.Render(w, err, httpStatus)
		return
	}

	metric, err := h.useCase.Execute(
		r.Context(),
		payload.Type,
		payload.Name,
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

var _ http.Handler = (*GetMetricV2Handler)(nil)

func NewGetMetricV2Handler(
	useCase GetMetricUseCase,
	errorPresenter presenter.ErrorPresenter,
	metricPresenter presenter.MetricPresenter,
	validate *validator.Validate,
) *GetMetricV2Handler {
	return &GetMetricV2Handler{
		useCase:         useCase,
		errorPresenter:  errorPresenter,
		metricPresenter: metricPresenter,
		validate:        validate,
	}
}
