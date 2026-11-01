package httpextra_test

import (
	"bytes"
	"context"
	"io"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
)

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

		for range batches {
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
