package impl

import (
	"errors"
	"math/rand/v2"
	"runtime"

	"github.com/nikimonax/go-metrics/internal/app"
	"github.com/nikimonax/go-metrics/internal/domain"
)

type CollectorFunc func() ([]domain.Metric, error)

func (f CollectorFunc) Collect() ([]domain.Metric, error) {
	return f()
}

type CollectorsGroup struct {
	collectors []app.MetricCollector
}

func (group *CollectorsGroup) Collect() ([]domain.Metric, error) {
	metrics := make([]domain.Metric, 0, len(group.collectors))
	errs := make([]error, 0, len(group.collectors))

	for _, collector := range group.collectors {
		metricsBatch, err := collector.Collect()

		if err == nil {
			metrics = append(metrics, metricsBatch...)
		} else {
			errs = append(errs, err)
		}
	}

	return metrics, errors.Join(errs...)
}

func NewCollectorsGroup(collectors ...app.MetricCollector) app.MetricCollector {
	return &CollectorsGroup{collectors: collectors}
}

func CollectMemStats() ([]domain.Metric, error) {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	return []domain.Metric{
		domain.NewGaugeMetric("Alloc", float64(stats.Alloc)),
		domain.NewGaugeMetric("BuckHashSys", float64(stats.BuckHashSys)),
		domain.NewGaugeMetric("Frees", float64(stats.Frees)),
		domain.NewGaugeMetric("GCCPUFraction", float64(stats.GCCPUFraction)),
		domain.NewGaugeMetric("GCSys", float64(stats.GCSys)),
		domain.NewGaugeMetric("HeapAlloc", float64(stats.HeapAlloc)),
		domain.NewGaugeMetric("HeapIdle", float64(stats.HeapIdle)),
		domain.NewGaugeMetric("HeapInuse", float64(stats.HeapInuse)),
		domain.NewGaugeMetric("HeapObjects", float64(stats.HeapObjects)),
		domain.NewGaugeMetric("HeapReleased", float64(stats.HeapReleased)),
		domain.NewGaugeMetric("HeapSys", float64(stats.HeapSys)),
		domain.NewGaugeMetric("LastGC", float64(stats.LastGC)),
		domain.NewGaugeMetric("Lookups", float64(stats.Lookups)),
		domain.NewGaugeMetric("MCacheInuse", float64(stats.MCacheInuse)),
		domain.NewGaugeMetric("MCacheSys", float64(stats.MCacheSys)),
		domain.NewGaugeMetric("MSpanInuse", float64(stats.MSpanInuse)),
		domain.NewGaugeMetric("MSpanSys", float64(stats.MSpanSys)),
		domain.NewGaugeMetric("Mallocs", float64(stats.Mallocs)),
		domain.NewGaugeMetric("NextGC", float64(stats.NextGC)),
		domain.NewGaugeMetric("NumForcedGC", float64(stats.NumForcedGC)),
		domain.NewGaugeMetric("NumGC", float64(stats.NumGC)),
		domain.NewGaugeMetric("OtherSys", float64(stats.OtherSys)),
		domain.NewGaugeMetric("PauseTotalNs", float64(stats.PauseTotalNs)),
		domain.NewGaugeMetric("StackInuse", float64(stats.StackInuse)),
		domain.NewGaugeMetric("StackSys", float64(stats.StackSys)),
		domain.NewGaugeMetric("Sys", float64(stats.Sys)),
		domain.NewGaugeMetric("TotalAlloc", float64(stats.TotalAlloc)),
	}, nil
}

func CollectRandomValue() ([]domain.Metric, error) {
	return []domain.Metric{
		domain.NewGaugeMetric("RandomValue", rand.Float64()),
	}, nil
}

func CollectIncrOne() ([]domain.Metric, error) {
	return []domain.Metric{
		domain.NewCounterMetric("PollCount", 1),
	}, nil
}
