package usecase

import (
	"context"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
)

// Сценарий восстановления метрик с диска
// и их загрузки в хранилище.
type RestoreMetricsUseCase struct {
	dumper     interfaces.MetricDumper
	repository interfaces.MetricRepository
}

// Execute загружает метрики с диска через
// дампер и обновляет их в хранилище.
func (useCase *RestoreMetricsUseCase) Execute(
	ctx context.Context,
) error {
	metrics, err := useCase.dumper.Load()

	if err != nil {
		return err
	}

	return useCase.repository.UpdateBatch(ctx, metrics)
}

// NewRestoreMetricsUseCase создаёт новый экземпляр RestoreMetricsUseCase
// с указанными дампером и хранилищем метрик.
func NewRestoreMetricsUseCase(
	dumper interfaces.MetricDumper,
	repository interfaces.MetricRepository,
) *RestoreMetricsUseCase {
	return &RestoreMetricsUseCase{
		dumper:     dumper,
		repository: repository,
	}
}
