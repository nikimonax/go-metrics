package impl

import (
	"github.com/nikimonax/go-metrics/internal/app"
	"github.com/nikimonax/go-metrics/internal/domain"
)

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

// Update implements [app.MetricRepository].
func (repo *InMemoryMetricRepository) Update(other domain.Metric) error {
	if metric, ok := repo.index.Find(other.Type(), other.Name()); ok {
		return metric.Accept(other)
	} else {
		return repo.index.Add(other)
	}
}

func NewInMemoryMetricRepository() app.MetricRepository {
	return &InMemoryMetricRepository{
		index: make(MetricIndex),
	}
}
