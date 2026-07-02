package internal

import (
	"github.com/nikimonax/go-metrics/pkg"
	"github.com/nikimonax/go-metrics/server/app"
)

type MetricIndex map[pkg.MetricType]map[pkg.MetricName]pkg.Metric

type InMemoryMetricRepository struct {
	index MetricIndex
}

func (index MetricIndex) Find(
	metricType pkg.MetricType,
	metricName pkg.MetricName,
) (metric pkg.Metric, ok bool) {
	metric, ok = index[metricType][metricName]
	return
}

func (index MetricIndex) Add(metric pkg.Metric) error {
	metricType := metric.Type()

	sub, ok := index[metricType]

	if !ok {
		sub = make(map[pkg.MetricName]pkg.Metric)
		index[metricType] = sub
	}

	sub[metric.Name()] = metric
	return nil
}

// Update implements [app.MetricRepository].
func (repo *InMemoryMetricRepository) Update(other pkg.Metric) error {
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
