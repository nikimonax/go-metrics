package usecase

import (
	"context"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
)

type SaveMetricsUseCase struct {
	dumper     interfaces.MetricDumper
	repository interfaces.MetricRepository
}

func (useCase *SaveMetricsUseCase) Execute(
	ctx context.Context,
) error {
	metrics, err := useCase.repository.GetAll(ctx)

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
