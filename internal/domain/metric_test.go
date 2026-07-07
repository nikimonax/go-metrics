package domain_test

import (
	"errors"
	"testing"

	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/mock"
	"github.com/stretchr/testify/assert"
)

func TestMetricType(t *testing.T) {
	type TestCase struct {
		name       string
		metricType string
		want       bool
	}

	tests := []TestCase{
		{
			name:       "valid counter",
			metricType: "counter",
			want:       true,
		},
		{
			name:       "valid gauge",
			metricType: "gauge",
			want:       true,
		},
		{
			name:       "invalid",
			metricType: "random",
			want:       false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			metricType := domain.MetricType(tc.metricType)
			got := metricType.IsValid()
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestMetricValue(t *testing.T) {
	type TestCase struct {
		name  string
		value domain.MetricValue
		want  string
	}

	tests := []TestCase{
		{
			name:  "counter",
			value: domain.CounterMetricValue(42),
			want:  "42",
		},
		{
			name:  "gauge",
			value: domain.GaugeMetricValue(3.14),
			want:  "3.140",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.value.String()
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestMetric_TypeNameValue(t *testing.T) {

	type TestCase struct {
		name      string
		metric    domain.Metric
		wantType  domain.MetricType
		wantName  domain.MetricName
		wantValue domain.MetricValue
	}

	tests := []TestCase{
		{
			name:      "counter",
			metric:    domain.NewCounterMetric("TestCounter", 42),
			wantType:  domain.Counter,
			wantName:  domain.MetricName("TestCounter"),
			wantValue: domain.CounterMetricValue(42),
		},
		{
			name:      "gauge",
			metric:    domain.NewGaugeMetric("TestGauge", 3.14),
			wantType:  domain.Gauge,
			wantName:  domain.MetricName("TestGauge"),
			wantValue: domain.GaugeMetricValue(3.14),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantType, tc.metric.Type())
			assert.Equal(t, tc.wantName, tc.metric.Name())
			assert.Equal(t, tc.wantValue, tc.metric.Value())
		})
	}
}

func TestMetric_Accept(t *testing.T) {
	type TestCase struct {
		name   string
		metric domain.Metric
		setup  func(*TestCase, *mock.MetricUpdater)
		err    error
	}

	tests := []TestCase{
		{
			name:   "counter success",
			metric: domain.NewCounterMetric("TestCounter", 42),
			setup: func(tc *TestCase, updater *mock.MetricUpdater) {
				updater.On("UpdateCounter", tc.metric).Return(nil)
			},
			err: nil,
		},
		{
			name:   "counter error",
			metric: domain.NewCounterMetric("TestCounter", 42),
			setup: func(tc *TestCase, updater *mock.MetricUpdater) {
				updater.On("UpdateCounter", tc.metric).Return(tc.err)
			},
			err: errors.New("counter accept error"),
		},
		{
			name:   "gauge success",
			metric: domain.NewGaugeMetric("TestGauge", 3.14),
			setup: func(tc *TestCase, updater *mock.MetricUpdater) {
				updater.On("UpdateGauge", tc.metric).Return(nil)
			},
			err: nil,
		},
		{
			name:   "gauge error",
			metric: domain.NewGaugeMetric("TestGauge", 3.14),
			setup: func(tc *TestCase, updater *mock.MetricUpdater) {
				updater.On("UpdateGauge", tc.metric).Return(tc.err)
			},
			err: errors.New("gauge accept error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			updater := new(mock.MetricUpdater)
			tc.setup(&tc, updater)

			err := tc.metric.Accept(updater)

			assert.ErrorIs(t, err, tc.err)
			updater.AssertExpectations(t)
		})
	}
}

func TestMetric_Update(t *testing.T) {
	type TestCase struct {
		name      string
		metricA   domain.Metric
		metricB   domain.Metric
		wantErr   error
		wantValue domain.MetricValue
	}

	tests := []TestCase{
		{
			name:      "counter counter",
			metricA:   domain.NewCounterMetric("A", 4),
			metricB:   domain.NewCounterMetric("B", 5),
			wantErr:   nil,
			wantValue: domain.CounterMetricValue(9),
		},
		{
			name:      "gauge gauge",
			metricA:   domain.NewGaugeMetric("A", 3.14),
			metricB:   domain.NewGaugeMetric("B", 2.71),
			wantErr:   nil,
			wantValue: domain.GaugeMetricValue(2.71),
		},
		{
			name:      "counter gauge",
			metricA:   domain.NewCounterMetric("A", 4),
			metricB:   domain.NewGaugeMetric("B", 2.71),
			wantErr:   domain.ErrUpdate,
			wantValue: nil,
		},
		{
			name:      "gauge counter",
			metricA:   domain.NewGaugeMetric("B", 2.71),
			metricB:   domain.NewCounterMetric("A", 4),
			wantErr:   domain.ErrUpdate,
			wantValue: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			metricBValueBefore := tc.metricB.Value()

			err := tc.metricA.Accept(tc.metricB)

			assert.ErrorIs(t, err, tc.wantErr)

			if err == nil {
				assert.Equal(t, tc.metricA.Value(), tc.wantValue)
				assert.Equal(t, tc.metricB.Value(), metricBValueBefore)
			}
		})
	}

}
