package mock

import (
	"github.com/nikimonax/go-metrics/internal/app"
	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MetricGateway struct {
	mock.Mock
}

// Send implements [app.MetricGateway].
func (gateway *MetricGateway) Send(metric domain.Metric) error {
	return gateway.Called(metric).Error(0)
}

// SendBatch implements [app.MetricGateway].
func (gateway *MetricGateway) SendBatch(metrics []domain.Metric) error {
	return gateway.Called(metrics).Error(0)
}

var _ app.MetricGateway = (*MetricGateway)(nil)
