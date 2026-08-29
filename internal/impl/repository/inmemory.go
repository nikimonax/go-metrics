package repository

import (
	"maps"
	"slices"
	"sync"

	"github.com/nikimonax/go-metrics/internal/app"
	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/domain"
)

type MetricIndex map[domain.MetricType]map[domain.MetricName]domain.Metric

// inmemory

type InMemoryMetricRepository struct {
	mu    sync.Mutex
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

func (repo *InMemoryMetricRepository) updateUnlocked(metric domain.Metric) error {
	if existing, ok := repo.index.Find(metric.Type(), metric.Name()); ok {
		return existing.Accept(metric)
	}
	return repo.index.Add(metric)
}

// Update implements [interfaces.MetricRepository].
func (repo *InMemoryMetricRepository) Update(metric domain.Metric) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	return repo.updateUnlocked(metric)
}

// UpdateBatch implements [interfaces.MetricRepository].
func (repo *InMemoryMetricRepository) UpdateBatch(metrics []domain.Metric) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	for _, metric := range metrics {
		if err := repo.updateUnlocked(metric); err != nil {
			return err
		}
	}
	return nil
}

// Get implements [interfaces.MetricRepository].
func (repo *InMemoryMetricRepository) Get(
	metricType domain.MetricType,
	metricName domain.MetricName,
) (domain.Metric, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	metric, ok := repo.index.Find(metricType, metricName)

	if !ok {
		return nil, app.NewErrMetricNotFound(metricType, metricName)
	}

	return metric, nil
}

// GetAll implements [interfaces.MetricRepository].
func (repo *InMemoryMetricRepository) GetAll() ([]domain.Metric, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	metrics := make([]domain.Metric, 0, repo.index.Len())

	for _, sub := range repo.index {
		metrics = append(metrics, slices.Collect(maps.Values(sub))...)
	}

	return metrics, nil
}

func (repo *InMemoryMetricRepository) Clear() error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	return repo.index.Clear()
}

func NewInMemoryMetricRepository() interfaces.MetricRepository {
	return &InMemoryMetricRepository{
		index: make(MetricIndex),
	}
}
