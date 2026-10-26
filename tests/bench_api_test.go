package tests

import (
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap/zapcore"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/impl/gateway"
	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	"github.com/nikimonax/go-metrics/internal/server/config"
	"github.com/nikimonax/go-metrics/internal/server/fxmodule"
)

func newTestServer(tb testing.TB) (*url.URL, func()) {
	tb.Helper()

	cfg := config.NewDefaultConfig()
	cfg.Listen = "127.0.0.1:0"
	cfg.Profile = true
	cfg.Log.Level = zapcore.ErrorLevel

	var router chi.Router

	app := fx.New(
		fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }),
		fx.Supply(&cfg),
		fxmodule.StorageModule(&cfg),
		fxmodule.CoreModule(),
		fxmodule.ProfileModule(&cfg),
		fxmodule.APIV1Module(),
		fxmodule.APIV2Module(),
		fxmodule.APIV3Module(),
		fx.Populate(
			fx.Annotate(
				&router,
				fx.ParamTags(`name:"router_base"`),
			),
		),
	)

	if err := app.Err(); err != nil {
		tb.Fatal(err)
	}

	if err := app.Start(context.Background()); err != nil {
		tb.Fatal(err)
	}

	server := httptest.NewServer(router)
	baseURL, err := url.Parse(server.URL)

	if err != nil {
		tb.Fatal(err)
	}

	return baseURL, func() {
		server.Close()
		_ = app.Stop(context.Background())
	}
}

func newTestGateway(tb testing.TB, version uint, baseURL *url.URL) interfaces.MetricGateway {
	tb.Helper()

	transport := http.DefaultTransport

	if version > 1 {
		transport = httpextra.NewCompressRoundTripper(transport, "gzip", gzip.BestSpeed)
	}

	factory, err := gateway.GetGatewayFactory(version)

	if err != nil {
		tb.Fatal(err)
	}

	client := &http.Client{Transport: transport}

	return factory(client, baseURL)
}

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
	baseURL, shutdown := newTestServer(b)
	defer shutdown()

	dataset := benchmarkDataset()

	for version := uint(1); version <= gateway.GetMaxAPIVersion(); version++ {
		name := fmt.Sprintf("v%d", version)
		b.Run(name, func(b *testing.B) {
			gw := newTestGateway(b, version, baseURL)

			for b.Loop() {
				if err := gw.SendBatch(b.Context(), dataset); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
