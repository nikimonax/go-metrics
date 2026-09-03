package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/domain"
)

type MetricRepository struct {
	mock.Mock
}

// Update implements [MetricRepository].
func (repo *MetricRepository) Update(
	ctx context.Context,
	metric domain.Metric,
) error {
	return repo.Called(ctx, metric).Error(0)
}

// UpdateBatch implements [MetricRepository].
func (repo *MetricRepository) UpdateBatch(
	ctx context.Context,
	metrics []domain.Metric,
) error {
	return repo.Called(ctx, metrics).Error(0)
}

// Get implements [MetricRepository].
func (repo *MetricRepository) Get(
	ctx context.Context,
	metricType domain.MetricType,
	metricName domain.MetricName,
) (domain.Metric, error) {
	args := repo.Called(ctx, metricType, metricName)
	err := args.Error(1)

	var metric domain.Metric

	if raw := args.Get(0); raw != nil {
		metric = raw.(domain.Metric)
	}

	return metric, err
}

// GetAll implements [MetricRepository].
func (repo *MetricRepository) GetAll(
	ctx context.Context,
) ([]domain.Metric, error) {
	args := repo.Called(ctx)
	return args.Get(0).([]domain.Metric), args.Error(1)
}

// PopAll implements [MetricRepository].
func (repo *MetricRepository) PopAll(
	ctx context.Context,
) ([]domain.Metric, error) {
	args := repo.Called(ctx)
	return args.Get(0).([]domain.Metric), args.Error(1)
}

// Clear implements [MetricRepository].
func (repo *MetricRepository) Clear(ctx context.Context) error {
	return repo.Called(ctx).Error(0)
}

var _ interfaces.MetricRepository = (*MetricRepository)(nil)
