package mock

import (
	"github.com/nikimonax/go-metrics/internal/app"
	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MetricCollector struct {
	mock.Mock
}

// Collect implements [app.MetricCollector].
func (collector *MetricCollector) Collect() ([]domain.Metric, error) {
	args := collector.Called()
	return args.Get(0).([]domain.Metric), args.Error(1)
}

var _ app.MetricCollector = (*MetricCollector)(nil)
