package domain

import (
	"errors"
	"fmt"
	"strconv"
)

type MetricType string
type MetricName string

const (
	Counter MetricType = "counter"
	Gauge   MetricType = "gauge"
)

var ErrUpdate = errors.New("failed update metric")

func (m MetricType) IsValid() bool {
	switch m {
	case Counter, Gauge:
		return true
	default:
		return false
	}
}

type MetricValue interface {
	fmt.Stringer
}

type MetricUpdater interface {
	UpdateCounter(m *CounterMetric) error
	UpdateGauge(m *GaugeMetric) error
}

type Metric interface {
	MetricUpdater
	Type() MetricType
	Name() MetricName
	Value() MetricValue
	Accept(MetricUpdater) error
}

// Counter

type CounterMetricValue int64

// String implements [fmt.Stringer].
func (v CounterMetricValue) String() string {
	return strconv.FormatInt(int64(v), 10)
}

type CounterMetric struct {
	name  MetricName
	value CounterMetricValue
}

// UpdateCounter implements [MetricUpdater].
func (m *CounterMetric) UpdateCounter(applyTo *CounterMetric) error {
	applyTo.value += m.value
	return nil
}

// UpdateGauge implements [MetricUpdater].
func (m *CounterMetric) UpdateGauge(applyTo *GaugeMetric) error {
	return fmt.Errorf(
		"%w: can't apply counter updater to the gauge metric",
		ErrUpdate,
	)
}

// Type implements [Metric].
func (m *CounterMetric) Type() MetricType {
	return Counter
}

// Name implements [Metric].
func (m *CounterMetric) Name() MetricName {
	return m.name
}

// Value implements [Metric].
func (m *CounterMetric) Value() MetricValue {
	return m.value
}

func (m *CounterMetric) InternalValue() int64 {
	return int64(m.value)
}

// Accept implements [Metric].
func (m *CounterMetric) Accept(u MetricUpdater) error {
	return u.UpdateCounter(m)
}

func NewCounterMetric(name MetricName, value int64) Metric {
	return &CounterMetric{
		name:  name,
		value: CounterMetricValue(value),
	}
}

// Gauge

type GaugeMetricValue float64

// String implements [fmt.Stringer].
func (v GaugeMetricValue) String() string {
	return strconv.FormatFloat(float64(v), 'f', -1, 64)
}

type GaugeMetric struct {
	name  MetricName
	value GaugeMetricValue
}

// UpdateCounter implements [MetricUpdater].
func (*GaugeMetric) UpdateCounter(m *CounterMetric) error {
	return fmt.Errorf(
		"%w: can't apply gauge updater to the counter metric",
		ErrUpdate,
	)
}

// UpdateGauge implements [MetricUpdater].
func (m *GaugeMetric) UpdateGauge(applyTo *GaugeMetric) error {
	applyTo.value = m.value
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

// Value implements [Metric].
func (m *GaugeMetric) Value() MetricValue {
	return m.value
}

func (m *GaugeMetric) InternalValue() float64 {
	return float64(m.value)
}

// Accept implements [Metric].
func (m *GaugeMetric) Accept(u MetricUpdater) error {
	return u.UpdateGauge(m)
}

func NewGaugeMetric(name MetricName, value float64) Metric {
	return &GaugeMetric{
		name:  name,
		value: GaugeMetricValue(value),
	}
}
