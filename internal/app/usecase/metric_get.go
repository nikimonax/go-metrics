package usecase

import (
	"context"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/domain"
)

type GetMetricUseCase struct {
	metricRepository interfaces.MetricRepository
}

func (useCase *GetMetricUseCase) Execute(
	ctx context.Context,
	metricType domain.MetricType,
	metricName domain.MetricName,
) (domain.Metric, error) {
	return useCase.metricRepository.Get(ctx, metricType, metricName)
}

func NewGetMetricUseCase(metricRepository interfaces.MetricRepository) *GetMetricUseCase {
	return &GetMetricUseCase{
		metricRepository: metricRepository,
	}
}
