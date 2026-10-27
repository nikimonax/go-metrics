package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	"github.com/nikimonax/go-metrics/internal/model"
	"github.com/nikimonax/go-metrics/internal/server/presenter"

	"github.com/go-playground/validator/v10"
)

type UpdateMetricV2Handler struct {
	useCase        UpdateMetricUseCase
	errorPresenter presenter.ErrorPresenter
	validate       *validator.Validate
}

// UpdateMetricV2Handler godoc
//
//	@ID				updateMetricV2
//	@Summary		Обновление метрики (JSON)
//	@Description	Обновляет значение метрики, переданной в теле запроса в формате JSON.
//	@Tags			metrics
//	@Accept			json
//	@Produce		json
//	@Param			metric	body		model.Metric	true	"Метрика"
//	@Success		200		{string}	string			"OK"
//	@Failure		400		{object}	model.ErrorResponse
//	@Failure		500		{object}	model.ErrorResponse
//	@Router			/update [post]
//
// ServeHTTP implements [http.Handler].
func (h *UpdateMetricV2Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var payload model.Metric

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

	metric := payload.ToDomain()

	if err := h.useCase.Execute(r.Context(), metric); err != nil {
		h.errorPresenter.Render(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set(httpextra.HDRContentType, httpextra.MIMEText)
	w.WriteHeader(http.StatusOK)
}

var _ http.Handler = (*UpdateMetricV2Handler)(nil)

func NewUpdateMetricV2Handler(
	useCase UpdateMetricUseCase,
	errorPresenter presenter.ErrorPresenter,
	validate *validator.Validate,
) *UpdateMetricV2Handler {
	return &UpdateMetricV2Handler{
		useCase:        useCase,
		errorPresenter: errorPresenter,
		validate:       validate,
	}
}
