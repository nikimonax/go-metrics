package tests

import (
	"fmt"
	"testing"

	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/impl/gateway"
	"github.com/nikimonax/go-metrics/internal/testing/shared"
)

func benchmarkDataset() []domain.Metric {
	names := []string{
		"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys",
		"HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased",
		"HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys",
		"MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC",
		"NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys",
		"Sys", "TotalAlloc", "RandomValue",
	}

	dataset := make([]domain.Metric, 0, len(names)+1)

	for i, name := range names {
		dataset = append(dataset, domain.NewGaugeMetric(domain.MetricName(name), float64(i)+0.25))
	}

	dataset = append(dataset, domain.NewCounterMetric(domain.MetricName("PollCount"), 1))

	return dataset
}

func BenchmarkAPI(b *testing.B) {
	baseURL, shutdown := shared.NewTestServer(b)
	defer shutdown()

	dataset := benchmarkDataset()

	for version := uint(1); version <= gateway.GetMaxAPIVersion(); version++ {
		name := fmt.Sprintf("v%d", version)
		b.Run(name, func(b *testing.B) {
			gw := shared.NewTestGateway(b, version, baseURL)

			for b.Loop() {
				if err := gw.SendBatch(b.Context(), dataset); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
