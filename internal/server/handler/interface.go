package handler

import (
	"context"

	"github.com/nikimonax/go-metrics/internal/domain"
)

type UpdateMetricUseCase interface {
	Execute(
		context.Context,
		domain.Metric,
	) error
}

type UpdateMetricsUseCase interface {
	Execute(
		context.Context,
		[]domain.Metric,
	) error
}

type GetMetricUseCase interface {
	Execute(
		context.Context,
		domain.MetricType,
		domain.MetricName,
	) (domain.Metric, error)
}

type GetAllMetricsUseCase interface {
	Execute(context.Context) ([]domain.Metric, error)
}
