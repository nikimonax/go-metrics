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

func NewUpdateMetricsV3Handler(
	useCase UpdateMetricsUseCase,
	errorPresenter presenter.ErrorPresenter,
	validate *validator.Validate,
) http.Handler {
	return &UpdateMetricsV3Handler{
		useCase:        useCase,
		errorPresenter: errorPresenter,
		validate:       validate,
	}
}
