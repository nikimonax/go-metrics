package mock

import (
	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/stretchr/testify/mock"
)

type Metric struct {
	mock.Mock
}

// UpdateCounter implements [domain.MetricUpdater].
func (m *Metric) UpdateCounter(applyTo *domain.CounterMetric) error {
	return m.Called(applyTo).Error(0)
}

// UpdateGauge implements [domain.MetricUpdater].
func (m *Metric) UpdateGauge(applyTo *domain.GaugeMetric) error {
	return m.Called(applyTo).Error(0)
}

// Type implements [domain.Metric].
func (m *Metric) Type() domain.MetricType {
	return m.Called().Get(0).(domain.MetricType)
}

// Name implements [domain.Metric].
func (m *Metric) Name() domain.MetricName {
	return m.Called().Get(0).(domain.MetricName)
}

// Value implements [domain.Metric].
func (m *Metric) Value() domain.MetricValue {
	return m.Called().Get(0).(domain.MetricValue)
}

// Accept implements [domain.Metric].
func (m *Metric) Accept(u domain.MetricUpdater) error {
	return m.Called(u).Error(0)
}

var _ domain.Metric = (*Metric)(nil)

type MetricUpdater struct {
	mock.Mock
}

// UpdateCounter implements [domain.MetricUpdater].
func (u *MetricUpdater) UpdateCounter(m *domain.CounterMetric) error {
	return u.Called(m).Error(0)
}

// UpdateGauge implements [domain.MetricUpdater].
func (u *MetricUpdater) UpdateGauge(m *domain.GaugeMetric) error {
	return u.Called(m).Error(0)
}

var _ domain.MetricUpdater = (*MetricUpdater)(nil)
