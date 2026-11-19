package handler

import (
	"context"
	"fmt"

	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/gen/metric"
)

func toDomainMetric(m *metric.Metric) (domain.Metric, error) {
	metricType := m.GetType()

	switch metricType {
	case metric.MetricType_METRIC_TYPE_GAUGE_UNSPECIFIED:
		return domain.NewGaugeMetric(
			domain.MetricName(m.GetId()),
			m.GetValue(),
		), nil
	case metric.MetricType_METRIC_TYPE_COUNTER:
		return domain.NewCounterMetric(
			domain.MetricName(m.GetId()),
			m.GetDelta(),
		), nil
	default:
		return nil, fmt.Errorf("unknown metric type '%s'", metricType.String())
	}
}

func toDomainMetrics(metrics []*metric.Metric) ([]domain.Metric, error) {
	result := make([]domain.Metric, 0, len(metrics))

	for _, m := range metrics {
		dm, err := toDomainMetric(m)

		if err != nil {
			return nil, err
		}

		result = append(result, dm)
	}

	return result, nil
}

type UpdateMetricsGrpcHandler struct {
	metric.UnimplementedMetricsServer

	updateMetrics UpdateMetricsUseCase
}

// UpdateMetrics implements [metric.MetricsServer].
func (handler *UpdateMetricsGrpcHandler) UpdateMetrics(
	ctx context.Context,
	req *metric.UpdateMetricsRequest,
) (*metric.UpdateMetricsResponse, error) {
	metrics, err := toDomainMetrics(req.GetMetrics())

	if err != nil {
		return nil, err
	}

	if err := handler.updateMetrics.Execute(ctx, metrics); err != nil {
		return nil, err
	}

	return new(metric.UpdateMetricsResponse), nil
}

var _ metric.MetricsServer = (*UpdateMetricsGrpcHandler)(nil)

func NewUpdateMetricsGrpcHandler(updateMetrics UpdateMetricsUseCase) *UpdateMetricsGrpcHandler {
	return &UpdateMetricsGrpcHandler{updateMetrics: updateMetrics}
}
