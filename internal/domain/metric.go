package domain

import "errors"

type MetricType string
type MetricName string

const (
	Counter MetricType = "counter"
	Gauge   MetricType = "gauge"
)

func (m MetricType) IsValid() bool {
	switch m {
	case Counter, Gauge:
		return true
	default:
		return false
	}
}

type MetricUpdater interface {
	UpdateCounter(m *CounterMetric) error
	UpdateGauge(m *GaugeMetric) error
}

type Metric interface {
	MetricUpdater
	Type() MetricType
	Name() MetricName
	Accept(MetricUpdater) error
}

// Counter

type CounterMetric struct {
	Value int64
	name  MetricName
}

// UpdateCounter implements [MetricUpdater].
func (m *CounterMetric) UpdateCounter(applyTo *CounterMetric) error {
	applyTo.Value += m.Value
	return nil
}

// UpdateGauge implements [MetricUpdater].
func (m *CounterMetric) UpdateGauge(applyTo *GaugeMetric) error {
	return errors.New("can't apply counter updater to the gauge metric")
}

// Type implements [Metric].
func (m *CounterMetric) Type() MetricType {
	return Counter
}

// Name implements [Metric].
func (m *CounterMetric) Name() MetricName {
	return m.name
}

// Accept implements [Metric].
func (m *CounterMetric) Accept(u MetricUpdater) error {
	return u.UpdateCounter(m)
}

func NewCounterMetric(name MetricName, value int64) Metric {
	return &CounterMetric{
		name:  name,
		Value: value,
	}
}

// Gauge

type GaugeMetric struct {
	Value float64
	name  MetricName
}

// UpdateCounter implements [MetricUpdater].
func (*GaugeMetric) UpdateCounter(m *CounterMetric) error {
	return errors.New("can't apply gauge updater to the counter metric")
}

// UpdateGauge implements [MetricUpdater].
func (m *GaugeMetric) UpdateGauge(applyTo *GaugeMetric) error {
	applyTo.Value = m.Value
	return nil
}

// Type implements [Metric].
func (m *GaugeMetric) Type() MetricType {
	return Gauge
}

// Name implements [Metric].
func (m *GaugeMetric) Name() MetricName {
	return m.name
}

// Accept implements [Metric].
func (m *GaugeMetric) Accept(u MetricUpdater) error {
	return u.UpdateGauge(m)
}

func NewGaugeMetric(name MetricName, value float64) Metric {
	return &GaugeMetric{
		name:  name,
		Value: value,
	}
}
