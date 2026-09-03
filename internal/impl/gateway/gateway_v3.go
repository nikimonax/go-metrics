package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	"github.com/nikimonax/go-metrics/internal/model"
)

type HTTPMetricV3Gateway struct {
	HTTPMetricV2Gateway
}

// Send implements [interfaces.MetricGateway].
func (gateway *HTTPMetricV3Gateway) Send(
	_ context.Context,
	_ domain.Metric,
) error {
	return errors.ErrUnsupported
}

// SendBatch implements [interfaces.MetricGateway].
func (gateway *HTTPMetricV3Gateway) SendBatch(
	ctx context.Context,
	metrics []domain.Metric,
) error {
	payload := make([]*model.Metric, 0, len(metrics))

	for _, metric := range metrics {
		payload = append(payload, model.NewMetricFromDomain(metric))
	}

	content, err := json.Marshal(payload)

	if err != nil {
		return fmt.Errorf("failed serialize metrics: %w", err)
	}

	if err := gateway.makeRequest(ctx, content); err != nil {
		return fmt.Errorf("failed send metrics: %w", err)
	}

	return nil
}

func NewHTTPMetricV3Gateway(baseURL *url.URL) interfaces.MetricGateway {
	return &HTTPMetricV3Gateway{
		HTTPMetricV2Gateway: HTTPMetricV2Gateway{
			endpoint: baseURL.JoinPath("updates").String() + "/",
			client: &http.Client{
				Transport: httpextra.NewCompressRoundTripper(
					http.DefaultTransport, "gzip",
				),
			},
			timeout: defaultRequestTimeout,
		},
	}
}
