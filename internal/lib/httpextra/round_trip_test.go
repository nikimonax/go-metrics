package httpextra_test

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	"github.com/nikimonax/go-metrics/internal/testing/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, errors.New("read error") }

func TestCompressRoundTripper(t *testing.T) {
	t.Run("success request", func(t *testing.T) {
		const payload = "request payload"
		var got *http.Request

		next := httpextra.RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
			got = r
			return &http.Response{
				StatusCode: http.StatusNoContent,
				Body:       io.NopCloser(strings.NewReader("")),
				Header:     make(http.Header),
			}, nil
		})

		req := shared.NewRequest(t, http.MethodPost, strings.NewReader(payload))
		resp, err := httpextra.NewCompressRoundTripper(next, httpextra.ENCGzip).RoundTrip(req)

		require.NoError(t, err)
		require.NotNil(t, resp)

		defer func() { assert.NoError(t, resp.Body.Close()) }()

		assert.Equal(t, httpextra.ENCGzip, got.Header.Get(httpextra.HDRContentEncoding))

		compressed, err := io.ReadAll(got.Body)
		require.NoError(t, err)

		assert.Equal(t, int64(len(compressed)), got.ContentLength)

		reader, err := gzip.NewReader(bytes.NewReader(compressed))
		require.NoError(t, err)

		decoded, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Equal(t, payload, string(decoded))
	})

	t.Run("transport error", func(t *testing.T) {
		var got *http.Request

		next := httpextra.RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
			got = r
			return nil, errors.New("transport error")
		})
		req := shared.NewRequest(t, http.MethodGet, nil)

		_, err := httpextra.NewCompressRoundTripper(next, httpextra.ENCGzip).RoundTrip(req) //nolint:bodyclose

		assert.Error(t, err)
		assert.Same(t, req, got)
	})

	t.Run("reader error", func(t *testing.T) {
		next := httpextra.RoundTripperFunc(func(_ *http.Request) (*http.Response, error) {
			t.Fatal("transport must not be called")
			return nil, nil
		})
		req := shared.NewRequest(t, http.MethodPost, errorReader{})
		req.ContentLength = -1

		_, err := httpextra.NewCompressRoundTripper(next, httpextra.ENCGzip).RoundTrip(req) //nolint:bodyclose

		assert.EqualError(t, err, "read error")
	})

	t.Run("unknown encoder", func(t *testing.T) {
		assert.PanicsWithValue(t, "unknown encoder 'br'", func() {
			httpextra.NewCompressRoundTripper(http.DefaultTransport, "br")
		})
	})
}
