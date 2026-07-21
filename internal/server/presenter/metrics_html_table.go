package presenter

import (
	"embed"
	"log"
	"net/http"
	"text/template"

	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	"go.uber.org/zap"
)

//go:embed templates/*.html
var templateFolder embed.FS

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

	if presenter.sugar == nil {
		return
	}

	presenter.sugar.Errorw(
		"failed to write http response",
		"err", err,
	)
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
