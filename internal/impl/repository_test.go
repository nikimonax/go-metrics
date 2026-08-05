package impl_test

import (
	"testing"

	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/impl"
	"github.com/stretchr/testify/require"
)

func TestInMemoryMetricRepository(t *testing.T) {
	repo := impl.NewInMemoryMetricRepository()

	metricA := domain.NewCounterMetric("A", 42)
	metricB := domain.NewGaugeMetric("B", 3.14)
	metricsAB := []domain.Metric{metricA, metricB}

	metrics, err := repo.GetAll()

	require.NoError(t, err)
	require.Empty(t, metrics)

	err = repo.UpdateBatch(metricsAB)

	require.NoError(t, err)

	metrics, err = repo.GetAll()

	require.NoError(t, err)
	require.ElementsMatch(t, metrics, metricsAB)

	_, err = repo.Get(domain.Counter, "C")

	require.ErrorIs(t, err, impl.ErrMetricNotFound)

	err = repo.Update(domain.NewCounterMetric("A", 1))

	require.NoError(t, err)

	metric, err := repo.Get(metricA.Type(), metricA.Name())

	require.NoError(t, err)
	require.Equal(t, "43", metric.Value().String())

	err = repo.Clear()

	require.NoError(t, err)

	metrics, err = repo.GetAll()

	require.NoError(t, err)
	require.Empty(t, metrics)
}
