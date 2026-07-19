package presenter

import (
	"net/http"

	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	"go.uber.org/zap"
)

type PlainTextMetricPresenter struct {
	sugar *zap.SugaredLogger
}

// RenderMetric implements [MetricPresenter].
func (presenter *PlainTextMetricPresenter) Render(
	w http.ResponseWriter, metric domain.Metric, code int,
) {
	w.Header().Set(httpextra.HDRContentType, httpextra.MIMEText)
	w.WriteHeader(code)

	_, err := w.Write([]byte(metric.Value().String()))

	if err == nil {
		return
	}

	if presenter.sugar != nil {
		presenter.sugar.Errorw(
			"failed to write http response",
			"err", err,
		)
	}
}

func NewPlainTextMetricPresenter(logger *zap.Logger) MetricPresenter {
	var sugar *zap.SugaredLogger

	if logger != nil {
		sugar = logger.Sugar()
	}

	return &PlainTextMetricPresenter{sugar: sugar}
}
