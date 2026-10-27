package usecase

import (
	"context"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
)

// Сценарий сохранения всех метрик из хранилища
// на диск через дампер.
type SaveMetricsUseCase struct {
	dumper     interfaces.MetricDumper
	repository interfaces.MetricRepository
}

// Execute получает все метрики из хранилища и сохраняет их на диск.
func (useCase *SaveMetricsUseCase) Execute(
	ctx context.Context,
) error {
	metrics, err := useCase.repository.GetAll(ctx)

	if err != nil {
		return err
	}

	return useCase.dumper.Save(metrics)
}

// NewSaveMetricsUseCase создаёт новый экземпляр SaveMetricsUseCase
// с указанными дампером и хранилищем метрик.
func NewSaveMetricsUseCase(
	dumper interfaces.MetricDumper,
	repository interfaces.MetricRepository,
) *SaveMetricsUseCase {
	return &SaveMetricsUseCase{
		dumper:     dumper,
		repository: repository,
	}
}
