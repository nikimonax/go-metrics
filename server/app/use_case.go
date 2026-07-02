package app

import (
	"github.com/nikimonax/go-metrics/pkg"
)

type UpdateMetricUseCase struct {
	metricRepository MetricRepository
}

func (useCase *UpdateMetricUseCase) Execute(metric pkg.Metric) error {
	return useCase.metricRepository.Update(metric)
}

func NewUpdateMetricUseCase(metricRepository MetricRepository) *UpdateMetricUseCase {
	return &UpdateMetricUseCase{
		metricRepository: metricRepository,
	}
}
