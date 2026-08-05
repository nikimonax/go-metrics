package gateway

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
)

const defaultRequestTimeout = time.Second

type HttpMetricGateway struct {
	baseUrl *url.URL
	client  *http.Client
	timeout time.Duration
}

// Send implements [interfaces.MetricGateway].
func (gateway *HttpMetricGateway) Send(metric domain.Metric) (err error) {
	url := gateway.baseUrl.JoinPath(
		"update",
		string(metric.Type()),
		string(metric.Name()),
		metric.Value().String(),
	).String()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		gateway.timeout,
	)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)

	if err != nil {
		return fmt.Errorf("failed create request: %w", err)
	}

	req.Header.Set(httpextra.HDRContentType, httpextra.MIMEText)

	resp, err := gateway.client.Do(req)

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

// SendBatch implements [interfaces.MetricGateway].
func (gateway *HttpMetricGateway) SendBatch(metrics []domain.Metric) error {
	for _, metric := range metrics {
		if err := gateway.Send(metric); err != nil {
			return err
		}
	}
	return nil
}

func NewHttpMetricGateway(baseUrl *url.URL) interfaces.MetricGateway {
	return &HttpMetricGateway{
		baseUrl: baseUrl,
		client:  &http.Client{},
		timeout: defaultRequestTimeout,
	}
}
