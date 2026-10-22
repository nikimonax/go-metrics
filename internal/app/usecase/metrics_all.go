package usecase

import (
	"context"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/domain"
)

// Сценарий получения всех метрик из хранилища.
type GetAllMetricsUseCase struct {
	metricRepository interfaces.MetricRepository
}

// Execute возвращает все метрики, находящиеся в хранилище.
func (useCase *GetAllMetricsUseCase) Execute(
	ctx context.Context,
) ([]domain.Metric, error) {
	return useCase.metricRepository.GetAll(ctx)
}

// NewGetAllMetricsUseCase создаёт новый экземпляр GetAllMetricsUseCase
// с указанным хранилищем метрик.
func NewGetAllMetricsUseCase(
	metricRepository interfaces.MetricRepository,
) *GetAllMetricsUseCase {
	return &GetAllMetricsUseCase{
		metricRepository: metricRepository,
	}
}
