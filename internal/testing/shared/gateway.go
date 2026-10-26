package shared

import (
	"compress/gzip"
	"net/http"
	"net/url"
	"testing"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/impl/gateway"
	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
)

func NewTestGateway(tb testing.TB, version uint, baseURL *url.URL) interfaces.MetricGateway {
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
