package usecase

import "github.com/nikimonax/go-metrics/internal/app/interfaces"

type RestoreMetricsUseCase struct {
	dumper     interfaces.MetricDumper
	repository interfaces.MetricRepository
}

func (useCase *RestoreMetricsUseCase) Execute() error {
	metrics, err := useCase.dumper.Load()

	if err != nil {
		return err
	}

	return useCase.repository.UpdateBatch(metrics)
}

func NewRestoreMetricsUseCase(
	dumper interfaces.MetricDumper,
	repository interfaces.MetricRepository,
) *RestoreMetricsUseCase {
	return &RestoreMetricsUseCase{
		dumper:     dumper,
		repository: repository,
	}
}
