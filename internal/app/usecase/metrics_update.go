package usecase

import (
	"context"

	"github.com/nikimonax/go-metrics/internal/app/event"
	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/domain"
)

type UpdateMetricsUseCase struct {
	metricRepository interfaces.MetricRepository
	eventPublisher   interfaces.EventPublisher
}

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

func NewUpdateMetricsUseCase(
	metricRepository interfaces.MetricRepository,
	eventPublisher interfaces.EventPublisher,
) *UpdateMetricsUseCase {
	return &UpdateMetricsUseCase{
		metricRepository: metricRepository,
		eventPublisher:   eventPublisher,
	}
}
