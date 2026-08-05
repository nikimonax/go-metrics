package usecase

import (
	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/domain"
)

type GetMetricUseCase struct {
	metricRepository interfaces.MetricRepository
}

func (useCase *GetMetricUseCase) Execute(
	metricType domain.MetricType,
	metricName domain.MetricName,
) (domain.Metric, error) {
	return useCase.metricRepository.Get(metricType, metricName)
}

func NewGetMetricUseCase(metricRepository interfaces.MetricRepository) *GetMetricUseCase {
	return &GetMetricUseCase{
		metricRepository: metricRepository,
	}
}
