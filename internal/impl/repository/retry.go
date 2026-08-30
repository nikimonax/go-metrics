package repository

import (
	"context"

	"github.com/sethvargo/go-retry"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/domain"
)

type RetryRepository struct {
	wrapped     interfaces.MetricRepository
	backoff     func() retry.Backoff
	isRetryable func(error) bool
}

// Update implements [interfaces.MetricRepository].
func (repo *RetryRepository) Update(
	ctx context.Context,
	metric domain.Metric,
) error {
	return retry.Do(ctx, repo.backoff(), func(ctx context.Context) error {
		err := repo.wrapped.Update(ctx, metric)

		if err != nil && repo.isRetryable(err) {
			err = retry.RetryableError(err)
		}

		return err
	})
}

// UpdateBatch implements [interfaces.MetricRepository].
func (repo *RetryRepository) UpdateBatch(
	ctx context.Context,
	metrics []domain.Metric,
) error {
	return retry.Do(ctx, repo.backoff(), func(ctx context.Context) error {
		err := repo.wrapped.UpdateBatch(ctx, metrics)

		if err != nil && repo.isRetryable(err) {
			err = retry.RetryableError(err)
		}

		return err
	})
}

// Get implements [interfaces.MetricRepository].
func (repo *RetryRepository) Get(
	ctx context.Context,
	metricType domain.MetricType,
	metricName domain.MetricName,
) (domain.Metric, error) {
	return retry.DoValue(
		ctx,
		repo.backoff(),
		func(ctx context.Context) (domain.Metric, error) {
			metric, err := repo.wrapped.Get(ctx, metricType, metricName)

			if err != nil && repo.isRetryable(err) {
				err = retry.RetryableError(err)
			}

			return metric, err
		},
	)
}

// GetAll implements [interfaces.MetricRepository].
func (repo *RetryRepository) GetAll(
	ctx context.Context,
) ([]domain.Metric, error) {
	return retry.DoValue(
		ctx,
		repo.backoff(),
		func(ctx context.Context) ([]domain.Metric, error) {
			metrics, err := repo.wrapped.GetAll(ctx)

			if err != nil && repo.isRetryable(err) {
				err = retry.RetryableError(err)
			}

			return metrics, err
		},
	)
}

// PopAll implements [interfaces.MetricRepository].
func (repo *RetryRepository) PopAll(
	ctx context.Context,
) ([]domain.Metric, error) {
	return retry.DoValue(
		ctx,
		repo.backoff(),
		func(ctx context.Context) ([]domain.Metric, error) {
			metrics, err := repo.wrapped.PopAll(ctx)

			if err != nil && repo.isRetryable(err) {
				err = retry.RetryableError(err)
			}

			return metrics, err
		},
	)
}

// Clear implements [interfaces.MetricRepository].
func (repo *RetryRepository) Clear(ctx context.Context) error {
	return retry.Do(
		ctx,
		repo.backoff(),
		func(ctx context.Context) error {
			err := repo.wrapped.Clear(ctx)

			if err != nil && repo.isRetryable(err) {
				err = retry.RetryableError(err)
			}

			return err
		},
	)
}

func NewRetryRepository(
	repo interfaces.MetricRepository,
	backoff func() retry.Backoff,
	isRetryable func(error) bool,
) interfaces.MetricRepository {
	return &RetryRepository{
		wrapped:     repo,
		backoff:     backoff,
		isRetryable: isRetryable,
	}
}
