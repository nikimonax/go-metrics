package usecase

import (
	"context"

	"github.com/nikimonax/go-metrics/internal/app/event"
	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/domain"
)

type UpdateMetricUseCase struct {
	metricRepository interfaces.MetricRepository
	eventPublisher   interfaces.EventPublisher
}

func (useCase *UpdateMetricUseCase) Execute(metric domain.Metric) error {
	if err := useCase.metricRepository.Update(metric); err != nil {
		return err
	}

	e := event.NewMetricsUpdatedEvent(metric)
	useCase.eventPublisher.Publish(context.TODO(), e)

	return nil
}

func NewUpdateMetricUseCase(
	metricRepository interfaces.MetricRepository,
	eventPublisher interfaces.EventPublisher,
) *UpdateMetricUseCase {
	return &UpdateMetricUseCase{
		metricRepository: metricRepository,
		eventPublisher:   eventPublisher,
	}
}
