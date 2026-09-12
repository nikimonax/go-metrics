package collector

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"runtime"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/domain"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

type CollectorFunc func() ([]domain.Metric, error)

func (f CollectorFunc) Collect() ([]domain.Metric, error) {
	return f()
}

type CollectorsGroup struct {
	collectors []interfaces.MetricCollector
}

func (group *CollectorsGroup) Collect() ([]domain.Metric, error) {
	var (
		metrics = make([]domain.Metric, 0, len(group.collectors))
		errs    error
	)

	for _, collector := range group.collectors {
		metricsBatch, err := collector.Collect()
		errs = errors.Join(errs, err)

		if err == nil {
			metrics = append(metrics, metricsBatch...)
		}
	}

	return metrics, errs
}

func NewCollectorsGroup(collectors ...interfaces.MetricCollector) interfaces.MetricCollector {
	return &CollectorsGroup{collectors: collectors}
}

func NewRuntimeStatsCollector() interfaces.MetricCollector {
	fn := func() ([]domain.Metric, error) {
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
	return CollectorFunc(fn)
}

func NewMemStatsCollector() interfaces.MetricCollector {
	fn := func() ([]domain.Metric, error) {
		vMem, err := mem.VirtualMemory()

		if err != nil {
			return nil, err
		}

		return []domain.Metric{
			domain.NewGaugeMetric("TotalMemory", float64(vMem.Total)),
			domain.NewGaugeMetric("FreeMemory", float64(vMem.Free)),
		}, nil

	}
	return CollectorFunc(fn)
}

func NewCPUStatsCollector() interfaces.MetricCollector {
	_, _ = cpu.Percent(0, true)

	fn := func() ([]domain.Metric, error) {
		usage, err := cpu.Percent(0, true)

		if err != nil {
			return nil, err
		}

		metrics := make([]domain.Metric, 0, len(usage))

		for i, u := range usage {
			name := fmt.Sprintf("CPUutilization%d", i+1)
			metric := domain.NewGaugeMetric(domain.MetricName(name), u)
			metrics = append(metrics, metric)
		}

		return metrics, nil
	}
	return CollectorFunc(fn)
}

func NewRandomGaugeCollector(name string) interfaces.MetricCollector {
	fn := func() ([]domain.Metric, error) {
		return []domain.Metric{
			domain.NewGaugeMetric(domain.MetricName(name), rand.Float64()),
		}, nil
	}
	return CollectorFunc(fn)
}

func NewCounterCollector(name string, inc int64) interfaces.MetricCollector {
	fn := func() ([]domain.Metric, error) {
		return []domain.Metric{
			domain.NewCounterMetric(domain.MetricName(name), inc),
		}, nil
	}
	return CollectorFunc(fn)
}
