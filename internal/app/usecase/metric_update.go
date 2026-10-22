package usecase

import (
	"context"

	"github.com/nikimonax/go-metrics/internal/app/event"
	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/domain"
)

// Сценарий обновления одной метрики в хранилище.
type UpdateMetricUseCase struct {
	metricRepository interfaces.MetricRepository
	eventPublisher   interfaces.EventPublisher
}

// Execute обновляет метрику в хранилище и после успешного обновления
// публикует событие event.MetricsUpdatedEvent.
func (useCase *UpdateMetricUseCase) Execute(
	ctx context.Context,
	metric domain.Metric,
) error {
	if err := useCase.metricRepository.Update(ctx, metric); err != nil {
		return err
	}

	e := event.NewMetricsUpdatedEvent(metric)
	useCase.eventPublisher.Publish(ctx, e)

	return nil
}

// NewUpdateMetricUseCase создаёт новый экземпляр UpdateMetricUseCase
// с указанными хранилищем метрик и издателем событий.
func NewUpdateMetricUseCase(
	metricRepository interfaces.MetricRepository,
	eventPublisher interfaces.EventPublisher,
) *UpdateMetricUseCase {
	return &UpdateMetricUseCase{
		metricRepository: metricRepository,
		eventPublisher:   eventPublisher,
	}
}
