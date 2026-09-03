package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/domain"
)

type MetricGateway struct {
	mock.Mock
}

// Send implements [interfaces.MetricGateway].
func (gateway *MetricGateway) Send(
	ctx context.Context,
	metric domain.Metric,
) error {
	return gateway.Called(ctx, metric).Error(0)
}

// SendBatch implements [interfaces.MetricGateway].
func (gateway *MetricGateway) SendBatch(
	ctx context.Context,
	metrics []domain.Metric,
) error {
	return gateway.Called(ctx, metrics).Error(0)
}

var _ interfaces.MetricGateway = (*MetricGateway)(nil)
