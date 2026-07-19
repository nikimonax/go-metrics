package mock

import (
	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/server/handler"
	"github.com/stretchr/testify/mock"
)

// update metric

type UpdateMetricUseCase struct {
	mock.Mock
}

// Execute implements [server.UpdateMetricUseCase].
func (useCase *UpdateMetricUseCase) Execute(metric domain.Metric) error {
	return useCase.Called(metric).Error(0)
}

var _ handler.UpdateMetricUseCase = (*UpdateMetricUseCase)(nil)

// get metric

type GetMetricUseCase struct {
	mock.Mock
}

// Execute implements [server.GetMetricUseCase].
func (useCase *GetMetricUseCase) Execute(
	metricType domain.MetricType,
	metricName domain.MetricName,
) (domain.Metric, error) {
	args := useCase.Called(metricType, metricName)
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
func (useCase *GetAllMetricsUseCase) Execute() ([]domain.Metric, error) {
	args := useCase.Called()

	metrics := make([]domain.Metric, 0)

	if raw := args.Get(0); raw != nil {
		metrics = raw.([]domain.Metric)
	}

	return metrics, args.Error(1)
}

var _ handler.GetAllMetricsUseCase = (*GetAllMetricsUseCase)(nil)
