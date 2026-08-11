package usecase

import "github.com/nikimonax/go-metrics/internal/app/interfaces"

type SaveMetricsUseCase struct {
	dumper     interfaces.MetricDumper
	repository interfaces.MetricRepository
}

func (useCase *SaveMetricsUseCase) Execute() error {
	metrics, err := useCase.repository.GetAll()

	if err != nil {
		return err
	}

	return useCase.dumper.Save(metrics)
}

func NewSaveMetricsUseCase(
	dumper interfaces.MetricDumper,
	repository interfaces.MetricRepository,
) *SaveMetricsUseCase {
	return &SaveMetricsUseCase{
		dumper:     dumper,
		repository: repository,
	}
}
