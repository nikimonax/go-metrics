package usecase

import (
	"context"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
)

// Сценарий сбора метрик с коллектора
// и их обновления в хранилище.
type CollectMetricsUseCase struct {
	collector  interfaces.MetricCollector
	repository interfaces.MetricRepository
}

// Execute собирает метрики через коллектор и обновляет их
// в хранилище. Если коллектор не вернул метрик, ничего не делает.
func (useCase *CollectMetricsUseCase) Execute(
	ctx context.Context,
) error {
	metrics, err := useCase.collector.Collect()

	if err != nil {
		return err
	}

	if len(metrics) == 0 {
		return nil
	}

	return useCase.repository.UpdateBatch(ctx, metrics)
}

// NewCollectMetricsUseCase создаёт новый экземпляр CollectMetricsUseCase
// с указанными коллектором и хранилищем метрик.
func NewCollectMetricsUseCase(
	collector interfaces.MetricCollector,
	repository interfaces.MetricRepository,
) *CollectMetricsUseCase {
	return &CollectMetricsUseCase{
		collector:  collector,
		repository: repository,
	}
}
