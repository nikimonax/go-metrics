package server

import (
	"embed"
	"html/template"
	"log"
	"net/http"

	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
)

//go:embed templates/*.html
var templateFolder embed.FS

// plain text error

type ErrorPresenter interface {
	Render(w http.ResponseWriter, err error, code int)
}

type MetricPresenter interface {
	Render(w http.ResponseWriter, metric domain.Metric, code int)
}

type MetricsPresenter interface {
	Render(w http.ResponseWriter, metrics []domain.Metric, code int)
}

// plain text metric

type PlainTextErrorPresenter struct{}

// RenderError implements [ErrorPresenter].
func (p *PlainTextErrorPresenter) Render(
	w http.ResponseWriter, err error, code int,
) {
	http.Error(w, err.Error(), code)
}

func NewPlainTextErrorPresenter() ErrorPresenter {
	return new(PlainTextErrorPresenter)
}

type PlainTextMetricPresenter struct{}

// RenderMetric implements [MetricPresenter].
func (p *PlainTextMetricPresenter) Render(
	w http.ResponseWriter, metric domain.Metric, code int,
) {
	w.Header().Set(httpextra.HDRContentType, httpextra.MIMEText)
	w.WriteHeader(code)

	if _, err := w.Write([]byte(metric.Value().String())); err != nil {
		log.Printf("Failed to write HTTP response: %v", err)
	}
}

func NewPlainTextMetricPresenter() MetricPresenter {
	return new(PlainTextMetricPresenter)
}

// html metrics table

type HtmlTableMetricsPresenter struct {
	pageTemplate *template.Template
}

// RenderMetrics implements [MetricsPresenter].
func (h *HtmlTableMetricsPresenter) Render(
	w http.ResponseWriter, metrics []domain.Metric, code int,
) {
	w.Header().Set(httpextra.HDRContentType, httpextra.MIMEHTML)
	w.WriteHeader(http.StatusOK)

	if err := h.pageTemplate.Execute(w, metrics); err != nil {
		log.Printf("Failed to write HTTP response: %v", err)
	}
}

func NewHtmlTableMetricsPresenter() MetricsPresenter {
	tmpl, err := template.ParseFS(templateFolder, "templates/metrics_table.html")

	if err != nil {
		log.Fatalf("Error parsing embedded template: %v", err)
	}

	return &HtmlTableMetricsPresenter{pageTemplate: tmpl}
}
