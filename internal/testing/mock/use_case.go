package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/server/handler"
)

// update metric

type UpdateMetricUseCase struct {
	mock.Mock
}

// Execute implements [server.UpdateMetricUseCase].
func (useCase *UpdateMetricUseCase) Execute(
	ctx context.Context,
	metric domain.Metric,
) error {
	return useCase.Called(ctx, metric).Error(0)
}

var _ handler.UpdateMetricUseCase = (*UpdateMetricUseCase)(nil)

// get metric

type GetMetricUseCase struct {
	mock.Mock
}

// Execute implements [server.GetMetricUseCase].
func (useCase *GetMetricUseCase) Execute(
	ctx context.Context,
	metricType domain.MetricType,
	metricName domain.MetricName,
) (domain.Metric, error) {
	args := useCase.Called(ctx, metricType, metricName)
	var metric domain.Metric

	if raw := args.Get(0); raw != nil {
		metric = raw.(domain.Metric)
	}

	return metric, args.Error(1)
}

var _ handler.GetMetricUseCase = (*GetMetricUseCase)(nil)

// get all metrics

type GetAllMetricsUseCase struct {
	mock.Mock
}

// Execute implements [server.GetAllMetricsUseCase].
func (useCase *GetAllMetricsUseCase) Execute(
	ctx context.Context,
) ([]domain.Metric, error) {
	args := useCase.Called(ctx)

	metrics := make([]domain.Metric, 0)

	if raw := args.Get(0); raw != nil {
		metrics = raw.([]domain.Metric)
	}

	return metrics, args.Error(1)
}

var _ handler.GetAllMetricsUseCase = (*GetAllMetricsUseCase)(nil)
