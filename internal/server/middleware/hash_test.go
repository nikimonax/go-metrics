package middleware_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/go-metrics/internal/server/middleware"
)

func TestVerifyHash(t *testing.T) {
	header := "HashSHA256"
	reqContent := "request"
	respContent := "response"
	key := "secret"

	t.Run("success", func(t *testing.T) {
		hashing := sha256.New
		h := hashing()
		_, err := h.Write([]byte(reqContent))
		require.NoError(t, err)
		_, err = h.Write([]byte(key))
		require.NoError(t, err)

		w := httptest.NewRecorder()
		r := httptest.NewRequestWithContext(
			t.Context(),
			http.MethodPost,
			"/test",
			bytes.NewReader([]byte(reqContent)),
		)
		r.Header.Add(header, hex.EncodeToString(h.Sum(nil)))

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			content, err := io.ReadAll(r.Body)
			require.NoError(t, err)

			defer func() { assert.NoError(t, r.Body.Close()) }()

			assert.Equal(t, reqContent, string(content))

			_, err = w.Write([]byte(respContent))
			assert.NoError(t, err)

			w.WriteHeader(http.StatusOK)
		})

		handler := middleware.VerifyHash(header, hashing, key)(nextHandler)
		handler.ServeHTTP(w, r)

		resp := w.Result()
		defer func() { assert.NoError(t, resp.Body.Close()) }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		content, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, respContent, string(content))
	})

	t.Run("missing header", func(t *testing.T) {
		called := false

		w := httptest.NewRecorder()
		r := httptest.NewRequestWithContext(
			t.Context(),
			http.MethodPost,
			"/test",
			bytes.NewReader([]byte(reqContent)),
		)

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})

		handler := middleware.VerifyHash(header, sha256.New, key)(nextHandler)
		handler.ServeHTTP(w, r)

		resp := w.Result()
		defer func() { assert.NoError(t, resp.Body.Close()) }()

		assert.False(t, called)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("wrong hash", func(t *testing.T) {
		called := false

		w := httptest.NewRecorder()
		r := httptest.NewRequestWithContext(
			t.Context(),
			http.MethodPost,
			"/test",
			bytes.NewReader([]byte(reqContent)),
		)
		r.Header.Set(header, "wrong")

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})

		handler := middleware.VerifyHash(header, sha256.New, key)(nextHandler)
		handler.ServeHTTP(w, r)

		resp := w.Result()
		defer func() { assert.NoError(t, resp.Body.Close()) }()

		assert.False(t, called)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
