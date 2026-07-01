package app

import (
	"github.com/nikimonax/go-metrics/pkg"
)

type MetricRepository interface {
	CounterAdd(metricName pkg.MetricName, value int64) error
	GaugeSet(metricName pkg.MetricName, value float64) error
}
