package interfaces

import (
	"context"

	"github.com/nikimonax/go-metrics/internal/domain"
)

type MetricRepository interface {
	Update(context.Context, domain.Metric) error
	UpdateBatch(context.Context, []domain.Metric) error
	Get(context.Context, domain.MetricType, domain.MetricName) (domain.Metric, error)
	GetAll(context.Context) ([]domain.Metric, error)
	PopAll(context.Context) ([]domain.Metric, error)
	Clear(context.Context) error
}

type MetricCollector interface {
	Collect() ([]domain.Metric, error)
}

type MetricGateway interface {
	Send(context.Context, domain.Metric) error
	SendBatch(context.Context, []domain.Metric) error
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
