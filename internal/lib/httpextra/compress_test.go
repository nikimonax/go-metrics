package httpextra

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/go-metrics/internal/lib/syncpool"
)

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, errors.New("read error") }

const (
	compressLevel = gzip.BestSpeed

	testURL = "http://example.com"
)

func TestCompressRoundTripper(t *testing.T) {
	t.Run("success request", func(t *testing.T) {
		const payload = "request payload"
		var got *http.Request

		next := RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
			got = r
			return &http.Response{
				StatusCode: http.StatusNoContent,
				Body:       io.NopCloser(strings.NewReader("")),
				Header:     make(http.Header),
			}, nil
		})

		rt, err := NewCompressRoundTripper(next, ENCGzip, compressLevel)
		require.NoError(t, err)

		req := httptest.NewRequestWithContext(
			t.Context(),
			http.MethodPost,
			testURL,
			strings.NewReader(payload),
		)

		resp, err := rt.RoundTrip(req)

		require.NoError(t, err)
		require.NotNil(t, resp)

		defer func() { assert.NoError(t, resp.Body.Close()) }()

		assert.Equal(t, ENCGzip, got.Header.Get(HDRContentEncoding))

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

		next := RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
			got = r
			return nil, errors.New("transport error")
		})

		rt, err := NewCompressRoundTripper(next, ENCGzip, compressLevel)
		require.NoError(t, err)

		req := httptest.NewRequestWithContext(
			t.Context(),
			http.MethodGet,
			testURL,
			nil,
		)

		resp, err := rt.RoundTrip(req)

		if err != nil && resp != nil {
			defer func() { assert.NoError(t, resp.Body.Close()) }()
		}

		assert.Error(t, err)
		assert.Same(t, req, got)
	})

	t.Run("reader error", func(t *testing.T) {
		next := RoundTripperFunc(func(_ *http.Request) (*http.Response, error) {
			t.Fatal("transport must not be called")
			return nil, nil
		})

		rt, err := NewCompressRoundTripper(next, ENCGzip, compressLevel)
		require.NoError(t, err)

		req := httptest.NewRequestWithContext(
			t.Context(),
			http.MethodPost,
			testURL,
			errorReader{},
		)
		req.ContentLength = -1

		resp, err := rt.RoundTrip(req)

		if err != nil && resp != nil {
			defer func() { assert.NoError(t, resp.Body.Close()) }()
		}

		assert.EqualError(t, err, "read error")
	})

	t.Run("skips already compressed request", func(t *testing.T) {
		const payload = "already compressed payload"

		var preCompressed bytes.Buffer
		gw := gzip.NewWriter(&preCompressed)
		_, err := gw.Write([]byte(payload))
		require.NoError(t, err)
		require.NoError(t, gw.Close())

		var got *http.Request
		next := RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
			got = r
			return &http.Response{
				StatusCode: http.StatusNoContent,
				Body:       io.NopCloser(strings.NewReader("")),
				Header:     make(http.Header),
			}, nil
		})

		rt, err := NewCompressRoundTripper(next, ENCGzip, compressLevel)
		require.NoError(t, err)

		req := httptest.NewRequestWithContext(
			t.Context(),
			http.MethodPost,
			testURL,
			bytes.NewReader(preCompressed.Bytes()),
		)
		req.Header.Set(HDRContentEncoding, ENCGzip)

		resp, err := rt.RoundTrip(req)

		require.NoError(t, err)
		require.NotNil(t, resp)

		defer func() { assert.NoError(t, resp.Body.Close()) }()

		assert.Equal(t, ENCGzip, got.Header.Get(HDRContentEncoding))
		assert.Equal(t, int64(preCompressed.Len()), got.ContentLength)

		body, err := io.ReadAll(got.Body)
		require.NoError(t, err)

		assert.Equal(t, preCompressed.Bytes(), body)
	})

	t.Run("unknown encoder", func(t *testing.T) {
		rt, err := NewCompressRoundTripper(http.DefaultTransport, "br", compressLevel)

		assert.Nil(t, rt)
		assert.Error(t, err)
	})
}

type passthroughWriter struct {
	w io.Writer
}

func (p *passthroughWriter) Write(b []byte) (int, error) {
	return p.w.Write(b)
}

func (p *passthroughWriter) Reset(w io.Writer) {
	p.w = w
}

func TestResetWriter(t *testing.T) {
	t.Run("does not pool buffer still in use", func(t *testing.T) {
		buffers := syncpool.New(func() *bytes.Buffer { return new(bytes.Buffer) })

		wb := newResetWriter(&passthroughWriter{w: io.Discard}, &buffers)
		wb.Reset()

		_, err := wb.Write([]byte("payload"))
		require.NoError(t, err)

		detached := wb.DetachBuffer()
		require.Equal(t, "payload", detached.String())

		reused := buffers.Get()

		assert.NotSame(t, detached, reused)
	})
}
