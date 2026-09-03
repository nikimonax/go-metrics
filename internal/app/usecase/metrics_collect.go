package usecase

import (
	"context"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
)

type CollectMetricsUseCase struct {
	collector  interfaces.MetricCollector
	repository interfaces.MetricRepository
}

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

func NewCollectMetricsUseCase(
	collector interfaces.MetricCollector,
	repository interfaces.MetricRepository,
) *CollectMetricsUseCase {
	return &CollectMetricsUseCase{
		collector:  collector,
		repository: repository,
	}
}
