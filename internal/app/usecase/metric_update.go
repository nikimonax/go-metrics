package usecase

import (
	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/domain"
)

type UpdateMetricUseCase struct {
	metricRepository interfaces.MetricRepository
}

func (useCase *UpdateMetricUseCase) Execute(metric domain.Metric) error {
	return useCase.metricRepository.Update(metric)
}

func NewUpdateMetricUseCase(metricRepository interfaces.MetricRepository) *UpdateMetricUseCase {
	return &UpdateMetricUseCase{
		metricRepository: metricRepository,
	}
}
