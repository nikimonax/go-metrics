package mock

import (
	"github.com/stretchr/testify/mock"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/domain"
)

type MetricCollector struct {
	mock.Mock
}

// Collect implements [interfaces.MetricCollector].
func (collector *MetricCollector) Collect() ([]domain.Metric, error) {
	args := collector.Called()
	return args.Get(0).([]domain.Metric), args.Error(1)
}

var _ interfaces.MetricCollector = (*MetricCollector)(nil)
