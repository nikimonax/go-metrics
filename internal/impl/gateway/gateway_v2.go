package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	"github.com/nikimonax/go-metrics/internal/model"
)

type HTTPMetricV2Gateway struct {
	endpoint string
	client   *http.Client
	timeout  time.Duration
}

func (gateway *HTTPMetricV2Gateway) makeRequest(
	ctx context.Context,
	content []byte,
) (err error) {
	ctx, cancel := context.WithTimeout(ctx, gateway.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		gateway.endpoint,
		bytes.NewBuffer(content),
	)

	if err != nil {
		return fmt.Errorf("failed create request: %w", err)
	}

	req.Header.Set(httpextra.HDRContentType, httpextra.MIMEJSON)

	resp, err := gateway.client.Do(req)

	if err != nil {
		return fmt.Errorf("request failed: %w", err)
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

		if resp.ContentLength > 0 {
			if body, err := io.ReadAll(resp.Body); err == nil {
				reason = string(body)
			}
		}

		return NewErrHTTPStatus(resp.StatusCode, reason)
	}

	return nil
}

// Send implements [interfaces.MetricGateway].
func (gateway *HTTPMetricV2Gateway) Send(
	ctx context.Context,
	metric domain.Metric,
) error {
	payload := model.NewMetricFromDomain(metric)

	content, err := json.Marshal(payload)

	if err != nil {
		return fmt.Errorf("failed serialize metric: %w", err)
	}

	if err := gateway.makeRequest(ctx, content); err != nil {
		return fmt.Errorf("failed send metric: %w", err)
	}

	return nil
}

// SendBatch implements [interfaces.MetricGateway].
func (gateway *HTTPMetricV2Gateway) SendBatch(
	ctx context.Context,
	metrics []domain.Metric,
) error {
	for _, metric := range metrics {
		if err := gateway.Send(ctx, metric); err != nil {
			return err
		}
	}
	return nil
}

func NewHTTPMetricV2Gateway(baseURL *url.URL) interfaces.MetricGateway {
	return &HTTPMetricV2Gateway{
		endpoint: baseURL.JoinPath("update").String() + "/",
		client: &http.Client{
			Transport: httpextra.NewCompressRoundTripper(
				http.DefaultTransport, "gzip",
			),
		},
		timeout: defaultRequestTimeout,
	}
}
