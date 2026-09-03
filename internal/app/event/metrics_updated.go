package event

import (
	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/domain"
)

var MetricsUpdatedEventName = "metrics.updated"

type MetricsUpdatedEvent struct {
	Metrics []domain.Metric
}

func (e MetricsUpdatedEvent) Name() string {
	return MetricsUpdatedEventName
}

var _ interfaces.Event = (*MetricsUpdatedEvent)(nil)

func NewMetricsUpdatedEvent(metrics ...domain.Metric) MetricsUpdatedEvent {
	return MetricsUpdatedEvent{Metrics: metrics}
}
