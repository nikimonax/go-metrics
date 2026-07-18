package server_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	"github.com/nikimonax/go-metrics/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlainTextErrorPresenter(t *testing.T) {
	message := "test error"
	status := http.StatusBadRequest
	err := errors.New(message)
	rr := httptest.NewRecorder()

	presenter := server.NewPlainTextErrorPresenter()
	presenter.Render(rr, err, status)

	resp := rr.Result()
	defer resp.Body.Close()

	contentType := resp.Header.Get(httpextra.HDRContentType)
	body, err := io.ReadAll(resp.Body)

	messageActual := strings.TrimSpace(string(body))

	require.NoError(t, err)
	assert.Equal(t, resp.StatusCode, status)
	assert.Contains(t, contentType, httpextra.MIMEText)
	assert.Equal(t, message, messageActual)
}

func TestPlainTextMetricPresenter(t *testing.T) {
	metric := domain.NewCounterMetric("TestMetric", 42)
	status := http.StatusOK
	rr := httptest.NewRecorder()

	presenter := server.NewPlainTextMetricPresenter(nil)
	presenter.Render(rr, metric, status)

	resp := rr.Result()
	defer resp.Body.Close()

	contentType := resp.Header.Get(httpextra.HDRContentType)
	body, err := io.ReadAll(resp.Body)

	require.NoError(t, err)
	assert.Equal(t, resp.StatusCode, status)
	assert.Contains(t, contentType, httpextra.MIMEText)
	assert.Equal(t, metric.Value().String(), string(body))
}

func TestHtmlTableMetricsPresenter(t *testing.T) {
	metrics := []domain.Metric{
		domain.NewCounterMetric("A", 42),
		domain.NewGaugeMetric("B", 3.14),
	}
	status := http.StatusOK

	rr := httptest.NewRecorder()

	presenter := server.NewHtmlTableMetricsPresenter(nil)
	presenter.Render(rr, metrics, status)

	resp := rr.Result()
	defer resp.Body.Close()

	contentType := resp.Header.Get(httpextra.HDRContentType)
	body, err := io.ReadAll(resp.Body)

	pageHtml := string(body)

	require.NoError(t, err)
	assert.Equal(t, resp.StatusCode, status)
	assert.Contains(t, contentType, httpextra.MIMEHTML)

	for _, metric := range metrics {
		assert.Contains(t, pageHtml, metric.Type())
		assert.Contains(t, pageHtml, metric.Name())
		assert.Contains(t, pageHtml, metric.Value().String())
	}
}
