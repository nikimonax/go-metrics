package gateway

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/gen/metric"
)

func metricFromDomain(m domain.Metric) (*metric.Metric, error) {
	builder := metric.Metric_builder{Id: string(m.Name())}

	switch v := m.(type) {
	case *domain.CounterMetric:
		builder.Type = metric.MetricType_METRIC_TYPE_COUNTER
		builder.Delta = v.InternalValue()
	case *domain.GaugeMetric:
		builder.Type = metric.MetricType_METRIC_TYPE_GAUGE_UNSPECIFIED
		builder.Value = v.InternalValue()
	default:
		return nil, fmt.Errorf("unknown metric type '%s'", m.Type())
	}

	return builder.Build(), nil
}

func requestFromDomain(metrics []domain.Metric) (*metric.UpdateMetricsRequest, error) {
	builder := metric.UpdateMetricsRequest_builder{
		Metrics: make([]*metric.Metric, 0, len(metrics)),
	}

	for _, d := range metrics {
		m, err := metricFromDomain(d)

		if err != nil {
			builder.Metrics = nil
			return nil, err
		}

		builder.Metrics = append(builder.Metrics, m)
	}

	return builder.Build(), nil
}

type GrpcMetricGateway struct {
	client metric.MetricsClient
}

// Send implements [interfaces.MetricGateway].
func (gateway *GrpcMetricGateway) Send(ctx context.Context, metric domain.Metric) error {
	return gateway.SendBatch(ctx, []domain.Metric{metric})
}

// SendBatch implements [interfaces.MetricGateway].
func (gateway *GrpcMetricGateway) SendBatch(ctx context.Context, metrics []domain.Metric) error {
	request, err := requestFromDomain(metrics)

	if err != nil {
		return err
	}

	_, err = gateway.client.UpdateMetrics(ctx, request)
	return err
}

var _ interfaces.MetricGateway = (*GrpcMetricGateway)(nil)

func NewGrpcMetricGateway(conn *grpc.ClientConn) *GrpcMetricGateway {
	return &GrpcMetricGateway{client: metric.NewMetricsClient(conn)}
}
