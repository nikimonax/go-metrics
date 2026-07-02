package app

import (
	"github.com/nikimonax/go-metrics/pkg"
)

type MetricRepository interface {
	Update(pkg.Metric) error
}
