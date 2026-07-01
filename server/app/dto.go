package app

import "github.com/nikimonax/go-metrics/pkg"

type UpdateMetricDTO struct {
	MetricType pkg.MetricType
	MetricName pkg.MetricName
	ValueAdd   int64
	ValueSet   float64
}
