package app

import (
	"github.com/nikimonax/go-metrics/internal/domain"
)

type MetricRepository interface {
	Update(domain.Metric) error
}
