package impl

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/nikimonax/go-metrics/internal/app"
	"github.com/nikimonax/go-metrics/internal/domain"
)

var ErrMetricNotFound = errors.New("metric not found")

type MetricIndex map[domain.MetricType]map[domain.MetricName]domain.Metric

type InMemoryMetricRepository struct {
	index MetricIndex
}

func (index MetricIndex) Find(
	metricType domain.MetricType,
	metricName domain.MetricName,
) (metric domain.Metric, ok bool) {
	metric, ok = index[metricType][metricName]
	return
}

func (index MetricIndex) Add(metric domain.Metric) error {
	metricType := metric.Type()

	sub, ok := index[metricType]

	if !ok {
		sub = make(map[domain.MetricName]domain.Metric)
		index[metricType] = sub
	}

	sub[metric.Name()] = metric
	return nil
}

func (index MetricIndex) Len() int {
	var totalLen int

	for _, sub := range index {
		totalLen += len(sub)
	}

	return totalLen
}

func (index MetricIndex) Clear() error {
	for _, sub := range index {
		clear(sub)
	}
	return nil
}

// Update implements [app.MetricRepository].
func (repo *InMemoryMetricRepository) Update(other domain.Metric) error {
	if metric, ok := repo.index.Find(other.Type(), other.Name()); ok {
		return metric.Accept(other)
	} else {
		return repo.index.Add(other)
	}
}

// UpdateBatch implements [app.MetricRepository].
func (repo *InMemoryMetricRepository) UpdateBatch(metrics []domain.Metric) error {
	for _, metric := range metrics {
		if err := repo.Update(metric); err != nil {
			return err
		}
	}
	return nil
}

// Get implements [app.MetricRepository].
func (repo *InMemoryMetricRepository) Get(
	metricType domain.MetricType,
	metricName domain.MetricName,
) (domain.Metric, error) {
	metric, ok := repo.index.Find(metricType, metricName)

	if !ok {
		return nil, fmt.Errorf("%w: (%s, %s)", ErrMetricNotFound, metricType, metricName)
	}

	return metric, nil
}

// GetAll implements [app.MetricRepository].
func (repo *InMemoryMetricRepository) GetAll() ([]domain.Metric, error) {
	metrics := make([]domain.Metric, 0, repo.index.Len())

	for _, sub := range repo.index {
		metrics = append(metrics, slices.Collect(maps.Values(sub))...)
	}

	return metrics, nil
}

func (repo *InMemoryMetricRepository) Clear() error {
	return repo.index.Clear()
}

func NewInMemoryMetricRepository() app.MetricRepository {
	return &InMemoryMetricRepository{
		index: make(MetricIndex),
	}
}
