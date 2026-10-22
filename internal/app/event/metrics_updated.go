// Пакет event содержит события application слоя,
// публикуемые через EventPublisher.

package event

import (
	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/domain"
)

const MetricsUpdatedEventName = "metrics.updated"

// Cобытие, возникающее при обновлении метрик.
type MetricsUpdatedEvent struct {
	// Список обновлённых метрик.
	Metrics []domain.Metric
}

// Name implements [interfaces.Event].
func (MetricsUpdatedEvent) Name() string {
	return MetricsUpdatedEventName
}

var _ interfaces.Event = (*MetricsUpdatedEvent)(nil)

func NewMetricsUpdatedEvent(metrics ...domain.Metric) MetricsUpdatedEvent {
	return MetricsUpdatedEvent{Metrics: metrics}
}
