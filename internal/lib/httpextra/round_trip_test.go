package httpextra_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	"github.com/nikimonax/go-metrics/internal/testing/shared"
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

		resp, err := httpextra.NewCompressRoundTripper(next, httpextra.ENCGzip).RoundTrip(req)

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
		req := shared.NewRequest(t, http.MethodPost, errorReader{})
		req.ContentLength = -1

		resp, err := httpextra.NewCompressRoundTripper(next, httpextra.ENCGzip).RoundTrip(req)

		if err != nil && resp != nil {
			defer func() { assert.NoError(t, resp.Body.Close()) }()
		}

		assert.EqualError(t, err, "read error")
	})

	t.Run("unknown encoder", func(t *testing.T) {
		assert.PanicsWithValue(t, "unknown encoder 'br'", func() {
			httpextra.NewCompressRoundTripper(http.DefaultTransport, "br")
		})
	})
}

func TestRateLimitRoundTripper(t *testing.T) {
	t.Run("verify limits", func(t *testing.T) {
		var (
			limit   = rand.Int()%10 + 1 // max rate limiter concurrency
			batches = rand.Int()%5 + 1
			jobs    = limit * batches

			ready   = make(chan struct{})
			release = make(chan struct{})

			cur atomic.Int64
		)

		require.Greater(t, limit, 0)
		require.Greater(t, batches, 0)

		next := httpextra.RoundTripperFunc(
			func(req *http.Request) (*http.Response, error) {
				assert.NoError(t, req.Body.Close())

				n := cur.Add(1)
				defer cur.Add(-1)

				if n > int64(limit) {
					assert.Fail(t, "concurrency exceeds limit")
				}

				ready <- struct{}{}
				<-release

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(nil)),
					Header:     make(http.Header),
				}, nil
			},
		)

		rt := httpextra.NewRateLimitRoundTripper(next, int64(limit))

		req := httptest.NewRequestWithContext(
			t.Context(),
			http.MethodGet,
			"/test",
			nil,
		)

		var wg sync.WaitGroup
		wg.Add(jobs)

		for range jobs {
			go func() {
				defer wg.Done()

				resp, err := rt.RoundTrip(req)
				require.NoError(t, err)
				assert.NoError(t, resp.Body.Close())
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			}()
		}

		for i := range batches {
			fmt.Print(i)
			for range limit {
				<-ready
			}
			for range limit {
				release <- struct{}{}
			}
		}

		done := make(chan struct{})

		go func() {
			wg.Wait()
			done <- struct{}{}
		}()

		select {
		case <-done:
		case <-time.After(time.Second):
			assert.Fail(t, "not all requests done")
		}
	})

	t.Run("test cancelable", func(t *testing.T) {
		release := make(chan struct{})

		next := httpextra.RoundTripperFunc(
			func(req *http.Request) (*http.Response, error) {
				assert.NoError(t, req.Body.Close())

				<-release

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(nil)),
					Header:     make(http.Header),
				}, nil
			},
		)

		rt := httpextra.NewRateLimitRoundTripper(next, 1)

		reqFirst := httptest.NewRequestWithContext(
			t.Context(),
			http.MethodGet,
			"/test",
			nil,
		)

		testCtx, cancel := context.WithCancel(context.Background())

		reqSecond := reqFirst.Clone(testCtx)

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			resp, err := rt.RoundTrip(reqFirst)
			require.NoError(t, err)
			assert.NoError(t, resp.Body.Close())
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}()

		go func() {
			defer wg.Done()
			resp, err := rt.RoundTrip(reqSecond)

			if err == nil {
				assert.NoError(t, resp.Body.Close())
			}

			require.ErrorIs(t, err, context.Canceled)
		}()

		done := make(chan struct{})

		go func() {
			wg.Wait()
			done <- struct{}{}
		}()

		cancel()
		release <- struct{}{}

		select {
		case <-done:
		case <-time.After(time.Second):
			assert.Fail(t, "not all requests done")
		}

	})
}
