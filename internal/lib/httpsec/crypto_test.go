package httpsec

import (
	"bytes"
	"crypto/rand"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	"github.com/nikimonax/go-metrics/internal/testing/mock"
)

const endpoint = "/test"

func generatePGPKey(t *testing.T) *crypto.Key {
	t.Helper()

	key, err := crypto.PGP().
		KeyGeneration().
		AddUserId("Test", "test@example.com").
		New().
		GenerateKey()
	require.NoError(t, err)

	return key
}

func encryptWithKey(t *testing.T, content []byte, key *crypto.Key) []byte {
	t.Helper()

	encHandle, err := crypto.PGP().Encryption().Recipient(key).New()
	require.NoError(t, err)

	message, err := encHandle.Encrypt(content)
	require.NoError(t, err)

	return message.Bytes()
}

func TestEncryptor(t *testing.T) {
	privateKey := generatePGPKey(t)

	encryptor := NewEncryptor()

	t.Run("encrypt decrypt compatibility", func(t *testing.T) {
		content := []byte(rand.Text())

		middleware, err := encryptor.DecryptMiddleware(privateKey)
		require.NoError(t, err)

		var handlerCalled = false
		handler := middleware(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true

				actualContent, err1 := io.ReadAll(r.Body)
				require.NoError(t, err1)

				assert.NoError(t, r.Body.Close())
				assert.Equal(t, content, actualContent)

				w.WriteHeader(http.StatusOK)
			}),
		)

		rt, err := encryptor.EncryptRoundTripper(
			httpextra.NewRoundTripperFromHandler(handler),
			privateKey,
		)
		require.NoError(t, err)

		body := mock.NewMockCloser(bytes.NewReader(content))
		req := httptest.NewRequestWithContext(
			t.Context(),
			http.MethodPost,
			endpoint,
			body,
		)

		resp, err := rt.RoundTrip(req)
		require.NoError(t, err)
		assert.NoError(t, resp.Body.Close())
		assert.True(t, body.Closed())
		assert.True(t, handlerCalled)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestDecryptMiddleware(t *testing.T) {
	privateKey := generatePGPKey(t)
	otherKey := generatePGPKey(t)

	encryptor := NewEncryptor()

	t.Run("pass through without header", func(t *testing.T) {
		content := []byte(rand.Text())

		middleware, err := encryptor.DecryptMiddleware(privateKey)
		require.NoError(t, err)

		var handlerCalled = false
		handler := middleware(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true

				actualContent, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				assert.NoError(t, r.Body.Close())
				assert.Equal(t, content, actualContent)

				w.WriteHeader(http.StatusOK)
			}),
		)

		body := mock.NewMockCloser(bytes.NewReader(content))
		r := httptest.NewRequestWithContext(
			t.Context(),
			http.MethodPost,
			endpoint,
			body,
		)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		assert.True(t, body.Closed())
		assert.True(t, handlerCalled)
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	})

	t.Run("pass through with false header", func(t *testing.T) {
		content := []byte(rand.Text())

		middleware, err := encryptor.DecryptMiddleware(privateKey)
		require.NoError(t, err)

		var handlerCalled = false
		handler := middleware(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true

				actualContent, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				assert.NoError(t, r.Body.Close())
				assert.Equal(t, content, actualContent)

				w.WriteHeader(http.StatusOK)
			}),
		)

		body := mock.NewMockCloser(bytes.NewReader(content))
		r := httptest.NewRequestWithContext(
			t.Context(),
			http.MethodPost,
			endpoint,
			body,
		)
		r.Header.Set("X-Encrypted", "false")

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		assert.True(t, body.Closed())
		assert.True(t, handlerCalled)
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	})

	t.Run("bad header value", func(t *testing.T) {
		middleware, err := encryptor.DecryptMiddleware(privateKey)
		require.NoError(t, err)

		var handlerCalled = false
		handler := middleware(
			http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			}),
		)

		r := httptest.NewRequestWithContext(
			t.Context(),
			http.MethodPost,
			endpoint,
			bytes.NewReader([]byte("content")),
		)
		r.Header.Set("X-Encrypted", "not-a-bool")

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		resp := w.Result()
		assert.NoError(t, resp.Body.Close())
		assert.False(t, handlerCalled)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("X-Content-Type-Options"), "nosniff")
	})

	t.Run("decrypt failure with wrong key", func(t *testing.T) {
		content := []byte(rand.Text())

		encrypted := encryptWithKey(t, content, otherKey)

		middleware, err := encryptor.DecryptMiddleware(privateKey)
		require.NoError(t, err)

		var handlerCalled = false
		handler := middleware(
			http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			}),
		)

		r := httptest.NewRequestWithContext(
			t.Context(),
			http.MethodPost,
			endpoint,
			bytes.NewReader(encrypted),
		)
		r.Header.Set("X-Encrypted", "true")

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		resp := w.Result()
		assert.NoError(t, resp.Body.Close())
		assert.False(t, handlerCalled)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("headers cleanup after decrypt", func(t *testing.T) {
		content := []byte(rand.Text())

		encrypted := encryptWithKey(t, content, privateKey)

		middleware, err := encryptor.DecryptMiddleware(privateKey)
		require.NoError(t, err)

		var handlerCalled = false
		handler := middleware(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true

				assert.Empty(t, r.Header.Get("X-Encrypted"))
				assert.Empty(t, r.Header.Get(httpextra.HDRContentLength))
				assert.Equal(t, int64(len(content)), r.ContentLength)

				actualContent, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				assert.NoError(t, r.Body.Close())
				assert.Equal(t, content, actualContent)

				w.WriteHeader(http.StatusOK)
			}),
		)

		r := httptest.NewRequestWithContext(
			t.Context(),
			http.MethodPost,
			endpoint,
			bytes.NewReader(encrypted),
		)
		r.Header.Set("X-Encrypted", "true")
		r.Header.Set(httpextra.HDRContentLength, strconv.Itoa(len(encrypted)))

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		assert.True(t, handlerCalled)
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	})
}

func TestEncryptRoundTripper(t *testing.T) {
	privateKey := generatePGPKey(t)

	encryptor := NewEncryptor()

	t.Run("empty body is not encrypted", func(t *testing.T) {
		var roundTripperCalled = false
		rt, err := encryptor.EncryptRoundTripper(
			httpextra.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
				roundTripperCalled = true

				assert.Empty(t, req.Header.Get("X-Encrypted"))
				assert.Nil(t, req.Body)

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       http.NoBody,
					Header:     make(http.Header),
				}, nil
			}),
			privateKey,
		)
		require.NoError(t, err)

		req := httptest.NewRequestWithContext(
			t.Context(),
			http.MethodGet,
			endpoint,
			nil,
		)

		resp, err := rt.RoundTrip(req)
		require.NoError(t, err)
		assert.NoError(t, resp.Body.Close())
		assert.True(t, roundTripperCalled)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("multiple request usage", func(t *testing.T) {
		content := []byte(rand.Text())
		expectedRespContent := []byte(rand.Text())

		decHandle, err := crypto.PGP().Decryption().DecryptionKey(privateKey).New()
		require.NoError(t, err)

		var roundTripperCalls = 0
		rt, err := encryptor.EncryptRoundTripper(
			httpextra.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
				roundTripperCalls++

				encrypted, err1 := io.ReadAll(req.Body)
				require.NoError(t, err1)

				assert.NoError(t, req.Body.Close())
				assert.Equal(t, "true", req.Header.Get("X-Encrypted"))
				assert.Equal(t, int64(len(encrypted)), req.ContentLength)

				decrypted, err1 := decHandle.Decrypt(encrypted, crypto.Bytes)
				require.NoError(t, err1)
				assert.Equal(t, content, decrypted.Bytes())

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(expectedRespContent)),
					Header:     make(http.Header),
				}, nil
			}),
			privateKey,
		)
		require.NoError(t, err)

		req := httptest.NewRequestWithContext(
			t.Context(),
			http.MethodPost,
			endpoint,
			nil,
		)

		for i := range 2 {
			body := mock.NewMockCloser(bytes.NewReader(content))
			req.Body = body
			req.ContentLength = int64(len(content))

			resp, err := rt.RoundTrip(req)
			require.NoError(t, err)

			respContent, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.NoError(t, resp.Body.Close())
			assert.Equal(t, expectedRespContent, respContent)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.True(t, body.Closed())
			assert.Equal(t, i+1, roundTripperCalls)
		}
	})

	t.Run("large body is encrypted", func(t *testing.T) {
		content := bytes.Repeat([]byte("a"), 4096)

		var roundTripperCalled = false
		rt, err := encryptor.EncryptRoundTripper(
			httpextra.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
				roundTripperCalled = true

				assert.Equal(t, "true", req.Header.Get("X-Encrypted"))

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       http.NoBody,
					Header:     make(http.Header),
				}, nil
			}),
			privateKey,
		)
		require.NoError(t, err)

		req := httptest.NewRequestWithContext(
			t.Context(),
			http.MethodPost,
			endpoint,
			bytes.NewReader(content),
		)

		resp, err := rt.RoundTrip(req)
		require.NoError(t, err)
		assert.NoError(t, resp.Body.Close())
		assert.True(t, roundTripperCalled)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
