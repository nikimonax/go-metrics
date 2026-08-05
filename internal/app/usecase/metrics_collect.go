package usecase

import "github.com/nikimonax/go-metrics/internal/app/interfaces"

type CollectMetricsUseCase struct {
	collector  interfaces.MetricCollector
	repository interfaces.MetricRepository
}

func (useCase *CollectMetricsUseCase) Execute() error {
	metrics, err := useCase.collector.Collect()

	if err != nil {
		return err
	}

	if len(metrics) == 0 {
		return nil
	}

	return useCase.repository.UpdateBatch(metrics)
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
