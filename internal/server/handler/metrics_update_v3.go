package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	"github.com/nikimonax/go-metrics/internal/model"
	"github.com/nikimonax/go-metrics/internal/server/presenter"

	"github.com/go-playground/validator/v10"
)

type UpdateMetricsV3Handler struct {
	useCase        UpdateMetricsUseCase
	errorPresenter presenter.ErrorPresenter
	validate       *validator.Validate
}

// UpdateMetricsV3Handler godoc
//
//	@ID				updateMetricsV3
//	@Summary		Пакетное обновление метрик (JSON)
//	@Description	Обновляет несколько метрик, переданных в теле запроса в формате JSON.
//	@Tags			metrics
//	@Accept			json
//	@Produce		json
//	@Param			metrics	body		[]model.Metric	true	"Список метрик"
//	@Success		200		{string}	string			"OK"
//	@Failure		400		{object}	model.ErrorResponse
//	@Failure		500		{object}	model.ErrorResponse
//	@Router			/updates [post]
//
// ServeHTTP implements [http.Handler].
func (h *UpdateMetricsV3Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	payload := make([]model.Metric, 0)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&payload); err != nil {
		h.errorPresenter.Render(w, err, http.StatusBadRequest)
		return
	}

	if err := h.validate.Var(payload, `required,dive`); err != nil {
		httpStatus := http.StatusInternalServerError

		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			httpStatus = http.StatusBadRequest
		}

		h.errorPresenter.Render(w, err, httpStatus)
		return
	}

	metrics := make([]domain.Metric, 0, len(payload))

	for _, m := range payload {
		metrics = append(metrics, m.ToDomain())
	}

	if err := h.useCase.Execute(r.Context(), metrics); err != nil {
		h.errorPresenter.Render(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set(httpextra.HDRContentType, httpextra.MIMEText)
	w.WriteHeader(http.StatusOK)
}

var _ http.Handler = (*UpdateMetricsV3Handler)(nil)

func NewUpdateMetricsV3Handler(
	useCase UpdateMetricsUseCase,
	errorPresenter presenter.ErrorPresenter,
	validate *validator.Validate,
) *UpdateMetricsV3Handler {
	return &UpdateMetricsV3Handler{
		useCase:        useCase,
		errorPresenter: errorPresenter,
		validate:       validate,
	}
}
