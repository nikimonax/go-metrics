package httpextra_test

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	"github.com/nikimonax/go-metrics/internal/testing/shared"
)

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, errors.New("read error") }

const compressLevel = gzip.BestSpeed

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

		rt, err := httpextra.NewCompressRoundTripper(next, httpextra.ENCGzip, compressLevel)
		require.NoError(t, err)

		req := shared.NewRequest(t, http.MethodPost, strings.NewReader(payload))
		resp, err := rt.RoundTrip(req)

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

		rt, err := httpextra.NewCompressRoundTripper(next, httpextra.ENCGzip, compressLevel)
		require.NoError(t, err)

		req := shared.NewRequest(t, http.MethodGet, nil)
		resp, err := rt.RoundTrip(req)

		if err != nil && resp != nil {
			defer func() { assert.NoError(t, resp.Body.Close()) }()
		}

		assert.Error(t, err)
		assert.Same(t, req, got)
	})

	t.Run("reader error", func(t *testing.T) {
		next := httpextra.RoundTripperFunc(func(_ *http.Request) (*http.Response, error) {
			t.Fatal("transport must not be called")
			return nil, nil
		})

		rt, err := httpextra.NewCompressRoundTripper(next, httpextra.ENCGzip, compressLevel)
		require.NoError(t, err)

		req := shared.NewRequest(t, http.MethodPost, errorReader{})
		req.ContentLength = -1

		resp, err := rt.RoundTrip(req)

		if err != nil && resp != nil {
			defer func() { assert.NoError(t, resp.Body.Close()) }()
		}

		assert.EqualError(t, err, "read error")
	})

	t.Run("unknown encoder", func(t *testing.T) {
		rt, err := httpextra.NewCompressRoundTripper(http.DefaultTransport, "br", compressLevel)

		assert.Nil(t, rt)
		assert.Error(t, err)
	})
}
