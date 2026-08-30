package usecase

import (
	"context"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/domain"
)

type GetAllMetricsUseCase struct {
	metricRepository interfaces.MetricRepository
}

func (useCase *GetAllMetricsUseCase) Execute(
	ctx context.Context,
) ([]domain.Metric, error) {
	return useCase.metricRepository.GetAll(ctx)
}

func NewGetAllMetricsUseCase(
	metricRepository interfaces.MetricRepository,
) *GetAllMetricsUseCase {
	return &GetAllMetricsUseCase{
		metricRepository: metricRepository,
	}
}
