package httpextra_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
)

func TestExtraHeadersRoundTripper(t *testing.T) {
	captureNext := func(t *testing.T, got *http.Header) httpextra.RoundTripperFunc {
		t.Helper()

		return httpextra.RoundTripperFunc(
			func(req *http.Request) (*http.Response, error) {
				*got = req.Header.Clone()

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(nil)),
					Header:     make(http.Header),
				}, nil
			},
		)
	}

	t.Run("sets configured headers", func(t *testing.T) {
		var got http.Header
		next := captureNext(t, &got)

		rt := httpextra.NewExtraHeadersRoundTripper(next, map[string]string{
			"X-Real-IP": "192.0.2.1",
			"X-Other":   "value",
		})

		req := httptest.NewRequestWithContext(
			t.Context(),
			http.MethodGet,
			"/test",
			nil,
		)

		resp, err := rt.RoundTrip(req)

		require.NoError(t, err)
		assert.NoError(t, resp.Body.Close())
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		assert.Equal(t, "192.0.2.1", got.Get("X-Real-IP"))
		assert.Equal(t, "value", got.Get("X-Other"))
	})

	t.Run("overwrites existing header values", func(t *testing.T) {
		var got http.Header
		next := captureNext(t, &got)

		rt := httpextra.NewExtraHeadersRoundTripper(
			next,
			map[string]string{"X-Real-IP": "192.0.2.1"},
		)

		req := httptest.NewRequestWithContext(
			t.Context(),
			http.MethodGet,
			"/test",
			nil,
		)
		req.Header.Set("X-Real-IP", "10.0.0.1")

		resp, err := rt.RoundTrip(req)

		require.NoError(t, err)
		assert.NoError(t, resp.Body.Close())
		assert.Equal(t, "192.0.2.1", got.Get("X-Real-IP"))
	})

	t.Run("forwards errors from next", func(t *testing.T) {
		wantErr := errors.New("round trip error")

		next := httpextra.RoundTripperFunc(
			func(*http.Request) (*http.Response, error) {
				return nil, wantErr
			},
		)

		rt := httpextra.NewExtraHeadersRoundTripper(
			next,
			map[string]string{"X-Real-IP": "192.0.2.1"},
		)

		req := httptest.NewRequestWithContext(
			t.Context(),
			http.MethodGet,
			"/test",
			nil,
		)

		resp, err := rt.RoundTrip(req)

		if err == nil {
			assert.NoError(t, resp.Body.Close())
		}

		require.ErrorIs(t, err, wantErr)
		assert.Nil(t, resp)
	})
}
