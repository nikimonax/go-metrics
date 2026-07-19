package handler

import "github.com/nikimonax/go-metrics/internal/domain"

type UpdateMetricUseCase interface {
	Execute(domain.Metric) error
}

type GetMetricUseCase interface {
	Execute(
		metricType domain.MetricType,
		metricName domain.MetricName,
	) (domain.Metric, error)
}

type GetAllMetricsUseCase interface {
	Execute() ([]domain.Metric, error)
}
