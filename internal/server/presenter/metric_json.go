package presenter

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	"github.com/nikimonax/go-metrics/internal/model"
)

type JSONMetricPresenter struct {
	sugar *zap.SugaredLogger
}

// RenderMetric implements [MetricPresenter].
func (presenter *JSONMetricPresenter) Render(
	w http.ResponseWriter, metric domain.Metric, code int,
) {
	w.Header().Set(httpextra.HDRContentType, httpextra.MIMEJSON)
	w.WriteHeader(code)

	payload := model.NewMetricFromDomain(metric)

	err := json.NewEncoder(w).Encode(payload)

	if err == nil {
		return
	}

	if presenter.sugar == nil {
		return
	}

	presenter.sugar.Errorw(
		"failed to write http response",
		"err", err,
	)
}

func NewJSONMetricPresenter(logger *zap.Logger) MetricPresenter {
	var sugar *zap.SugaredLogger

	if logger != nil {
		sugar = logger.Sugar()
	}

	return &JSONMetricPresenter{sugar: sugar}
}
