package usecase

import (
	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/domain"
)

type GetAllMetricsUseCase struct {
	metricRepository interfaces.MetricRepository
}

func (useCase *GetAllMetricsUseCase) Execute() ([]domain.Metric, error) {
	return useCase.metricRepository.GetAll()
}

func NewGetAllMetricsUseCase(metricRepository interfaces.MetricRepository) *GetAllMetricsUseCase {
	return &GetAllMetricsUseCase{
		metricRepository: metricRepository,
	}
}
