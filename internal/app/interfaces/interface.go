package interfaces

import (
	"context"

	"github.com/nikimonax/go-metrics/internal/domain"
)

type MetricRepository interface {
	Update(domain.Metric) error
	UpdateBatch([]domain.Metric) error
	Get(domain.MetricType, domain.MetricName) (domain.Metric, error)
	GetAll() ([]domain.Metric, error)
	PopAll() ([]domain.Metric, error)
	Clear() error
}

type MetricCollector interface {
	Collect() ([]domain.Metric, error)
}

type MetricGateway interface {
	Send(domain.Metric) error
	SendBatch([]domain.Metric) error
}

type MetricDumper interface {
	Save([]domain.Metric) error
	Load() ([]domain.Metric, error)
}

type Event interface {
	Name() string
}

type EventPublisher interface {
	Publish(context.Context, Event)
}
