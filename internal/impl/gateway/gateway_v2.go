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

type HttpMetricV2Gateway struct {
	endpoint string
	client   *http.Client
	timeout  time.Duration
}

// Send implements [interfaces.MetricGateway].
func (gateway *HttpMetricV2Gateway) Send(metric domain.Metric) (err error) {
	payload := model.NewMetricFromDomain(metric)

	content, err := json.Marshal(payload)

	if err != nil {
		return fmt.Errorf("failed serialize metric: %w", err)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		gateway.timeout,
	)
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
func (gateway *HttpMetricV2Gateway) SendBatch(metrics []domain.Metric) error {
	for _, metric := range metrics {
		if err := gateway.Send(metric); err != nil {
			return err
		}
	}
	return nil
}

func NewHttpMetricV2Gateway(baseUrl *url.URL) interfaces.MetricGateway {
	return &HttpMetricV2Gateway{
		endpoint: baseUrl.JoinPath("update").String() + "/",
		client: &http.Client{
			Transport: httpextra.NewCompressRoundTripper(
				http.DefaultTransport, "gzip",
			),
		},
		timeout: defaultRequestTimeout,
	}
}
