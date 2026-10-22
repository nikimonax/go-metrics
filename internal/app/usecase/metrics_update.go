package usecase

import (
	"context"

	"github.com/nikimonax/go-metrics/internal/app/event"
	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/domain"
)

// Сценарий пакетного обновления метрик в хранилище.
type UpdateMetricsUseCase struct {
	metricRepository interfaces.MetricRepository
	eventPublisher   interfaces.EventPublisher
}

// Execute пакетно обновляет метрики в хранилище и после успешного
// обновления публикует событие event.MetricsUpdatedEvent.
// Если список метрик пуст, ничего не делает.
func (useCase *UpdateMetricsUseCase) Execute(
	ctx context.Context,
	metrics []domain.Metric,
) error {
	if len(metrics) == 0 {
		return nil
	}

	if err := useCase.metricRepository.UpdateBatch(ctx, metrics); err != nil {
		return err
	}

	e := event.NewMetricsUpdatedEvent(metrics...)
	useCase.eventPublisher.Publish(ctx, e)

	return nil
}

// NewUpdateMetricsUseCase создаёт новый экземпляр UpdateMetricsUseCase
// с указанными хранилищем метрик и издателем событий.
func NewUpdateMetricsUseCase(
	metricRepository interfaces.MetricRepository,
	eventPublisher interfaces.EventPublisher,
) *UpdateMetricsUseCase {
	return &UpdateMetricsUseCase{
		metricRepository: metricRepository,
		eventPublisher:   eventPublisher,
	}
}
