package gateway

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"syscall"

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

func NewRetryGateway(
	gateway interfaces.MetricGateway,
	backoff func() retry.Backoff,
	isRetryable func(error) bool,
) interfaces.MetricGateway {
	return &RetryGateway{
		wrapped:     gateway,
		backoff:     backoff,
		isRetryable: isRetryable,
	}
}

func HTTPErrorIsRetryable(err error) bool {
	if err == nil {
		return false
	}

	var statusErr *HTTPStatusError
	if errors.As(err, &statusErr) {
		switch statusErr.StatusCode {
		case http.StatusTooManyRequests, // 429
			http.StatusInternalServerError, // 500
			http.StatusBadGateway,          // 502
			http.StatusServiceUnavailable,  // 503
			http.StatusGatewayTimeout:      // 504
			return true
		}

		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	switch {
	case errors.Is(err, io.EOF),
		errors.Is(err, io.ErrUnexpectedEOF),
		errors.Is(err, syscall.ECONNRESET),
		errors.Is(err, syscall.ECONNREFUSED),
		errors.Is(err, syscall.ECONNABORTED),
		errors.Is(err, syscall.EPIPE),
		errors.Is(err, syscall.EHOSTUNREACH),
		errors.Is(err, syscall.ENETUNREACH),
		errors.Is(err, syscall.ENETDOWN):
		return true
	}

	return false
}
