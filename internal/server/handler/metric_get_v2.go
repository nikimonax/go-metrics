package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/nikimonax/go-metrics/internal/impl"
	"github.com/nikimonax/go-metrics/internal/model"
	"github.com/nikimonax/go-metrics/internal/server/presenter"
)

type GetMetricV2Handler struct {
	useCase         GetMetricUseCase
	errorPresenter  presenter.ErrorPresenter
	metricPresenter presenter.MetricPresenter
	validate        *validator.Validate
}

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

	metric, err := h.useCase.Execute(payload.Type, payload.Name)

	if err != nil {
		status := http.StatusInternalServerError

		if errors.Is(err, impl.ErrMetricNotFound) {
			status = http.StatusNotFound
		}

		h.errorPresenter.Render(w, err, status)
		return

	}

	h.metricPresenter.Render(w, metric, http.StatusOK)
}

func NewGetMetricV2Handler(
	useCase GetMetricUseCase,
	errorPresenter presenter.ErrorPresenter,
	metricPresenter presenter.MetricPresenter,
) http.Handler {
	return &GetMetricV2Handler{
		useCase:         useCase,
		errorPresenter:  errorPresenter,
		metricPresenter: metricPresenter,
		validate:        getValidate(),
	}
}
