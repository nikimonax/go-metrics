package impl_test

import (
	"errors"
	"reflect"
	"runtime"
	"testing"

	"github.com/nikimonax/go-metrics/internal/app"
	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/impl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectorsGroup(t *testing.T) {
	type TestCase struct {
		name        string
		collectors  []app.MetricCollector
		wantMetrics []domain.Metric
		wantErrs    []error
	}

	metricA := domain.NewCounterMetric("A", 42)
	metricB := domain.NewGaugeMetric("B", 3.14)
	someErr := errors.New("some error")

	tests := []TestCase{
		{
			name:        "empty",
			collectors:  make([]app.MetricCollector, 0),
			wantMetrics: make([]domain.Metric, 0),
			wantErrs:    make([]error, 0),
		},
		{
			name: "success",
			collectors: []app.MetricCollector{
				impl.CollectorFunc(func() ([]domain.Metric, error) {
					return []domain.Metric{metricA}, nil
				}),
				impl.CollectorFunc(func() ([]domain.Metric, error) {
					return []domain.Metric{metricB}, nil
				}),
			},
			wantMetrics: []domain.Metric{metricA, metricB},
			wantErrs:    make([]error, 0),
		},
		{
			name: "error",
			collectors: []app.MetricCollector{
				impl.CollectorFunc(func() ([]domain.Metric, error) {
					return []domain.Metric{metricA, metricB}, nil
				}),
				impl.CollectorFunc(func() ([]domain.Metric, error) {
					return []domain.Metric{}, someErr
				}),
			},
			wantMetrics: []domain.Metric{metricA, metricB},
			wantErrs:    []error{someErr},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			collector := impl.NewCollectorsGroup(tc.collectors...)
			metrics, err := collector.Collect()

			assert.ElementsMatch(t, metrics, tc.wantMetrics)

			if len(tc.wantErrs) == 0 {
				assert.NoError(t, err)
			} else {
				for _, wantErr := range tc.wantErrs {
					assert.ErrorIs(t, err, wantErr)
				}
			}
		})
	}
}

func TestCollectMemStats(t *testing.T) {
	metrics, err := impl.CollectMemStats()

	require.NoError(t, err)
	require.NotEmpty(t, metrics)

	var stats runtime.MemStats
	statsType := reflect.TypeOf(stats)

	for _, metric := range metrics {
		assert.Equal(t, metric.Type(), domain.Gauge)
		_, ok := statsType.FieldByName(string(metric.Name()))
		assert.True(t, ok)
	}
}

func TestCollectRandomValue(t *testing.T) {
	metrics, err := impl.CollectRandomValue()

	require.NoError(t, err)
	require.Len(t, metrics, 1)

	metric := metrics[0]

	require.Equal(t, metric.Type(), domain.Gauge)
	require.Equal(t, metric.Name(), domain.MetricName("RandomValue"))
}

func TestCollectIncrOne(t *testing.T) {
	metrics, err := impl.CollectIncrOne()

	require.NoError(t, err)
	require.Len(t, metrics, 1)

	metric := metrics[0]

	require.Equal(t, metric.Type(), domain.Counter)
	require.Equal(t, metric.Name(), domain.MetricName("PollCount"))
	require.IsType(t, domain.CounterMetricValue(0), metric.Value())
	require.Equal(t, int64(metric.Value().(domain.CounterMetricValue)), int64(1))
}
