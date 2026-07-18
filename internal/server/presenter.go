package server

import (
	"embed"
	"html/template"
	"log"
	"net/http"

	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/lib/httpextra"

	"go.uber.org/zap"
)

//go:embed templates/*.html
var templateFolder embed.FS

type ErrorPresenter interface {
	Render(w http.ResponseWriter, err error, code int)
}

type MetricPresenter interface {
	Render(w http.ResponseWriter, metric domain.Metric, code int)
}

type MetricsPresenter interface {
	Render(w http.ResponseWriter, metrics []domain.Metric, code int)
}

// plain text error

type PlainTextErrorPresenter struct{}

// RenderError implements [ErrorPresenter].
func (presenter *PlainTextErrorPresenter) Render(
	w http.ResponseWriter, err error, code int,
) {
	http.Error(w, err.Error(), code)
}

func NewPlainTextErrorPresenter() ErrorPresenter {
	return new(PlainTextErrorPresenter)
}

// plain text metric

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

// html metrics table

type HtmlTableMetricsPresenter struct {
	pageTemplate *template.Template
	sugar        *zap.SugaredLogger
}

// RenderMetrics implements [MetricsPresenter].
func (presenter *HtmlTableMetricsPresenter) Render(
	w http.ResponseWriter, metrics []domain.Metric, code int,
) {
	w.Header().Set(httpextra.HDRContentType, httpextra.MIMEHTML)
	w.WriteHeader(http.StatusOK)

	err := presenter.pageTemplate.Execute(w, metrics)

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

func NewHtmlTableMetricsPresenter(logger *zap.Logger) MetricsPresenter {
	tmpl, err := template.ParseFS(templateFolder, "templates/metrics_table.html")

	if err != nil {
		log.Fatalf("failed parse embedded template: %v", err)
	}

	var sugar *zap.SugaredLogger

	if logger != nil {
		sugar = logger.Sugar()
	}

	return &HtmlTableMetricsPresenter{
		pageTemplate: tmpl,
		sugar:        sugar,
	}
}
