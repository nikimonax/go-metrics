package gateway

import (
	"context"

	"github.com/sethvargo/go-retry"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/domain"
)

type RetryGateway struct {
	wrapped     interfaces.MetricGateway
	backoff     func() retry.Backoff
	isRetryable func(error) bool
}

// Send implements [interfaces.MetricGateway].
func (gateway *RetryGateway) Send(
	ctx context.Context,
	metric domain.Metric,
) error {
	return retry.Do(ctx, gateway.backoff(), func(ctx context.Context) error {
		err := gateway.wrapped.Send(ctx, metric)

		if err != nil && gateway.isRetryable(err) {
			err = retry.RetryableError(err)
		}

		return err
	})
}

// SendBatch implements [interfaces.MetricGateway].
func (gateway *RetryGateway) SendBatch(
	ctx context.Context,
	metrics []domain.Metric,
) error {
	return retry.Do(ctx, gateway.backoff(), func(ctx context.Context) error {
		err := gateway.wrapped.SendBatch(ctx, metrics)

		if err != nil && gateway.isRetryable(err) {
			err = retry.RetryableError(err)
		}

		return err
	})
}

var _ interfaces.MetricGateway = (*RetryGateway)(nil)

func NewRetryGateway(
	gateway interfaces.MetricGateway,
	backoff func() retry.Backoff,
	isRetryable func(error) bool,
) *RetryGateway {
	return &RetryGateway{
		wrapped:     gateway,
		backoff:     backoff,
		isRetryable: isRetryable,
	}
}
