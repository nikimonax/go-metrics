package app

import (
	"github.com/nikimonax/go-metrics/internal/domain"
)

type MetricRepository interface {
	Update(domain.Metric) error
	UpdateBatch([]domain.Metric) error
	GetAll() ([]domain.Metric, error)
}

type MetricCollector interface {
	Collect() ([]domain.Metric, error)
}

type MetricGateway interface {
	Send(domain.Metric) error
	SendBatch([]domain.Metric) error
}
