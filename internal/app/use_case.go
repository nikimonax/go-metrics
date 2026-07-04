package app

import (
	"github.com/nikimonax/go-metrics/internal/domain"
)

type UpdateMetricUseCase struct {
	metricRepository MetricRepository
}

func (useCase *UpdateMetricUseCase) Execute(metric domain.Metric) error {
	return useCase.metricRepository.Update(metric)
}

func NewUpdateMetricUseCase(metricRepository MetricRepository) *UpdateMetricUseCase {
	return &UpdateMetricUseCase{
		metricRepository: metricRepository,
	}
}
