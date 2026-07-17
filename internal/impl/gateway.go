package impl

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/nikimonax/go-metrics/internal/app"
	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
)

type HttpMetricGateway struct {
	baseUrl *url.URL
	client  *http.Client
}

// Send implements [app.MetricGateway].
func (gateway *HttpMetricGateway) Send(metric domain.Metric) (err error) {
	url := gateway.baseUrl.JoinPath(
		"update",
		string(metric.Type()),
		string(metric.Name()),
		metric.Value().String(),
	).String()

	resp, err := gateway.client.Post(url, httpextra.MIMEText, nil) // nolint:noctx

	if err != nil {
		return fmt.Errorf("failed send metric: %w", err)
	}

	defer func() {
		if _, deferErr := io.Copy(io.Discard, resp.Body); deferErr != nil {
			err = errors.Join(err, deferErr)
		}

		if deferErr := resp.Body.Close(); deferErr != nil {
			err = errors.Join(err, deferErr)
		}
	}()

	if resp.StatusCode >= 400 {
		reason := "unknown"

		if resp.Header.Get(httpextra.HDRContentType) == httpextra.MIMEText {
			if body, err := io.ReadAll(resp.Body); err == nil {
				reason = string(body)
			}
		}

		return fmt.Errorf(
			"failed send metric (%d): %s",
			resp.StatusCode,
			reason,
		)
	}

	return nil
}

// SendBatch implements [app.MetricGateway].
func (gateway *HttpMetricGateway) SendBatch(metrics []domain.Metric) error {
	for _, metric := range metrics {
		if err := gateway.Send(metric); err != nil {
			return err
		}
	}
	return nil
}

func NewHttpMetricGateway(baseUrl *url.URL) app.MetricGateway {
	return &HttpMetricGateway{
		baseUrl: baseUrl,
		client:  &http.Client{},
	}
}
