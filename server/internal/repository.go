package internal

import (
	"fmt"

	"github.com/nikimonax/go-metrics/pkg"
	"github.com/nikimonax/go-metrics/server/app"
)

type InMemoryMetricRepository struct {
	gauges   map[pkg.MetricName]float64
	counters map[pkg.MetricName]int64
}

// CounterAdd implements [app.MetricRepository].
func (repo *InMemoryMetricRepository) CounterAdd(metricName pkg.MetricName, value int64) error {
	repo.counters[metricName] += value
	fmt.Println(*repo) // для отладки
	return nil
}

// GaugeSet implements [app.MetricRepository].
func (repo *InMemoryMetricRepository) GaugeSet(metricName pkg.MetricName, value float64) error {
	repo.gauges[metricName] = value
	fmt.Println(*repo) // для отладки
	return nil
}

func NewInMemoryMetricRepository() app.MetricRepository {
	return &InMemoryMetricRepository{
		gauges:   make(map[pkg.MetricName]float64),
		counters: make(map[pkg.MetricName]int64),
	}
}
