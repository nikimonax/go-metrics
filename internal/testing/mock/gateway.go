package mock

import (
	"github.com/stretchr/testify/mock"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/domain"
)

type MetricGateway struct {
	mock.Mock
}

// Send implements [interfaces.MetricGateway].
func (gateway *MetricGateway) Send(metric domain.Metric) error {
	return gateway.Called(metric).Error(0)
}

// SendBatch implements [interfaces.MetricGateway].
func (gateway *MetricGateway) SendBatch(metrics []domain.Metric) error {
	return gateway.Called(metrics).Error(0)
}

var _ interfaces.MetricGateway = (*MetricGateway)(nil)
