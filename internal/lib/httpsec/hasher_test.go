package httpsec_test

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	"github.com/nikimonax/go-metrics/internal/lib/httpsec"
	"github.com/nikimonax/go-metrics/internal/testing/mock"
)

const (
	headerKey = "HashSHA256"
	secretKey = "secret"

	endpoint = "/test"

	defaultContentStr = "Hello, World!"
	defaultSignature  = "fcfaffa7fef86515c7beb6b62d779fa4ccf092f2e61c164376054271252821ff"
)

var (
	defaultContent = []byte(defaultContentStr)

	hasher = httpsec.NewHasher(headerKey, sha256.New, []byte(secretKey))
)

func assertFuncsSame(t *testing.T, f1 any, f2 any) bool {
	ptr1 := reflect.ValueOf(f1).Pointer()
	ptr2 := reflect.ValueOf(f2).Pointer()
	return assert.Equal(t, ptr1, ptr2)

}

func TestHasher(t *testing.T) {
	t.Run("sign and verify", func(t *testing.T) {
		content := []byte(defaultContent)
		signature := defaultSignature

		newSignature := hasher.Sign(content)
		assert.Equal(t, signature, newSignature)
		assert.True(t, hasher.Verify(content, signature))

		badSignature := signature[1:]
		assert.False(t, hasher.Verify(content, badSignature))

		badSignature = "g" + badSignature
		assert.False(t, hasher.Verify(content, badSignature))
	})

	t.Run("calculate middleware returns handler", func(t *testing.T) {
		assertFuncsSame(t, hasher.CalculateHandler, hasher.CalculateMiddleware())
	})
}

func TestHasherVerifyHandler(t *testing.T) {
	type TestCase struct {
		name                string
		method              string
		content             []byte
		signature           string
		requireHeader       bool
		expectHandlerCalled bool
		expectStatusCode    int
	}

	tests := []TestCase{
		{
			name:                "empty body",
			method:              http.MethodGet,
			content:             []byte{},
			signature:           "",
			requireHeader:       true,
			expectHandlerCalled: true,
			expectStatusCode:    http.StatusOK,
		},
		{
			name:                "body with correct hash",
			method:              http.MethodPost,
			content:             defaultContent,
			signature:           defaultSignature,
			requireHeader:       true,
			expectHandlerCalled: true,
			expectStatusCode:    http.StatusOK,
		},
		{
			name:                "empty body with hash header",
			method:              http.MethodGet,
			content:             []byte{},
			signature:           defaultSignature,
			requireHeader:       true,
			expectHandlerCalled: false,
			expectStatusCode:    http.StatusBadRequest,
		},
		{
			name:                "body without hash header",
			method:              http.MethodPost,
			content:             defaultContent,
			signature:           "",
			requireHeader:       true,
			expectHandlerCalled: false,
			expectStatusCode:    http.StatusBadRequest,
		},
		{
			name:                "body with wrong hash 1",
			method:              http.MethodPost,
			content:             defaultContent,
			signature:           "wrong",
			requireHeader:       true,
			expectHandlerCalled: false,
			expectStatusCode:    http.StatusBadRequest,
		},
		{
			name:                "body with wrong hash 2",
			method:              http.MethodPost,
			content:             defaultContent,
			signature:           "a" + defaultSignature[1:],
			requireHeader:       true,
			expectHandlerCalled: false,
			expectStatusCode:    http.StatusBadRequest,
		},
		{
			name:                "body without require header",
			method:              http.MethodPost,
			content:             defaultContent,
			signature:           "",
			requireHeader:       false,
			expectHandlerCalled: false,
			expectStatusCode:    http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			respContent := []byte("response content")

			opts := make([]httpsec.VerifyOption, 0)

			if tc.requireHeader {
				opts = append(opts, httpsec.RequireHeader())
			}

			var handlerCalled = false
			handler := hasher.VerifyMiddleware(opts...)(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					handlerCalled = true

					received, err := io.ReadAll(r.Body)

					require.NoError(t, err)
					assert.NoError(t, r.Body.Close())
					assert.Equal(t, tc.content, received)

					w.WriteHeader(http.StatusOK)
					_, err = w.Write(respContent)
					assert.NoError(t, err)
				}),
			)

			body := mock.NewMockCloser(bytes.NewReader(tc.content))

			w := httptest.NewRecorder()
			r := httptest.NewRequestWithContext(
				t.Context(),
				tc.method,
				endpoint,
				body,
			)

			if tc.signature != "" {
				r.Header.Set(headerKey, tc.signature)
			}

			handler.ServeHTTP(w, r)

			resp := w.Result()
			actualRespContent, err := io.ReadAll(resp.Body)

			require.NoError(t, err)

			assert.NoError(t, resp.Body.Close())
			assert.True(t, body.Closed())

			if resp.StatusCode == http.StatusOK {
				assert.Equal(t, respContent, actualRespContent)
			}

			if tc.expectHandlerCalled {
				assert.True(t, handlerCalled)
			}

			if tc.expectStatusCode > 0 {
				assert.Equal(t, tc.expectStatusCode, resp.StatusCode)
			}
		})
	}
}

func TestHasherCalculateHandler(t *testing.T) {
	type TestCase struct {
		name               string
		respPanic          bool
		respStatusCode     int
		respContent        []byte
		expectedStatusCode int
		expectedSignature  string
	}

	tests := []TestCase{
		{
			name:               "empty body",
			respStatusCode:     http.StatusOK,
			respContent:        []byte{},
			expectedStatusCode: http.StatusOK,
			expectedSignature:  "",
		},
		{
			name:               "not empty body",
			respStatusCode:     http.StatusOK,
			respContent:        defaultContent,
			expectedStatusCode: http.StatusOK,
			expectedSignature:  defaultSignature,
		},
		{
			name:               "not empty body with 4xx error",
			respStatusCode:     http.StatusTeapot,
			respContent:        defaultContent,
			expectedStatusCode: http.StatusTeapot,
			expectedSignature:  defaultSignature,
		},
		{
			name:               "handle panic",
			respPanic:          true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedSignature:  "*",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := hasher.CalculateHandler(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.NoError(t, r.Body.Close())

					if tc.respPanic {
						panic("panic")
					}

					w.WriteHeader(tc.respStatusCode)
					_, err := w.Write(tc.respContent)
					assert.NoError(t, err)
				}),
			)

			w := httptest.NewRecorder()
			r := httptest.NewRequestWithContext(
				t.Context(),
				http.MethodGet,
				endpoint,
				nil,
			)

			if tc.respPanic {
				assert.Panics(t, func() {
					handler.ServeHTTP(w, r)
				})
				return
			}

			handler.ServeHTTP(w, r)

			resp := w.Result()
			respContent, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.NoError(t, resp.Body.Close())

			actualSignature := resp.Header.Get(headerKey)

			if tc.respContent != nil {
				assert.Equal(t, tc.respContent, respContent)
			}

			if tc.expectedSignature == "*" {
				assert.NotEmpty(t, actualSignature)
			} else {
				assert.Equal(t, tc.expectedSignature, actualSignature)
			}

			if tc.expectedStatusCode > 0 {
				assert.Equal(t, tc.expectedStatusCode, resp.StatusCode)
			}
		})
	}
}

func TestHasherRoundTripper(t *testing.T) {
	type TestCase struct {
		name              string
		method            string
		content           []byte
		expectedSignature string
	}

	tests := []TestCase{
		{
			name:              "empty body",
			method:            http.MethodGet,
			content:           []byte{},
			expectedSignature: "",
		},
		{
			name:              "not empty body",
			method:            http.MethodGet,
			content:           defaultContent,
			expectedSignature: defaultSignature,
		},
	}

	for _, tc := range tests {
		name := fmt.Sprintf("multiple request usage + %s", tc.name)

		t.Run(name, func(t *testing.T) {
			expectedStatusCode := http.StatusOK
			expectedRespContent := []byte(rand.Text())

			var roundTripperCalls = 0
			rt := hasher.RoundTripper(httpextra.RoundTripperFunc(
				func(req *http.Request) (*http.Response, error) {
					roundTripperCalls++

					content, err := io.ReadAll(req.Body)
					require.NoError(t, err)

					assert.NoError(t, req.Body.Close())
					assert.Equal(t, tc.content, content)
					assert.Equal(t, tc.expectedSignature, req.Header.Get(headerKey))

					return &http.Response{
						StatusCode: expectedStatusCode,
						Body:       io.NopCloser(bytes.NewReader(expectedRespContent)),
						Header:     make(http.Header),
					}, nil

				},
			))

			req := httptest.NewRequestWithContext(
				t.Context(),
				http.MethodPost,
				endpoint,
				nil,
			)

			for i := range 2 {
				body := mock.NewMockCloser(bytes.NewReader(tc.content))

				req.Body = body
				req.ContentLength = int64(len(tc.content))

				resp, err := rt.RoundTrip(req)
				require.NoError(t, err)

				respContent, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				assert.NoError(t, resp.Body.Close())
				assert.Equal(t, expectedRespContent, respContent)
				assert.Equal(t, expectedStatusCode, resp.StatusCode)
				assert.True(t, body.Closed())
				assert.Equal(t, roundTripperCalls, i+1)
			}
		})
	}

	t.Run("verify middleware compatibility", func(t *testing.T) {
		content := []byte(rand.Text())

		var handlerCalled = false
		handler := hasher.VerifyMiddleware(httpsec.RequireHeader())(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true

				actualContent, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				assert.NoError(t, r.Body.Close())
				assert.Equal(t, content, actualContent)

				w.WriteHeader(http.StatusOK)
			}),
		)

		rt := hasher.RoundTripper(httpextra.NewRoundTripperFromHandler(handler))

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
