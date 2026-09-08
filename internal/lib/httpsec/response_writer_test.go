package httpsec_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/go-metrics/internal/lib/httpsec"
)

func TestResponseWriterWithHashHeader(t *testing.T) {
	const (
		xHeader = "X-Custom"
		xValue  = "value"
	)

	type TestCase struct {
		name      string
		content   []byte
		signature string
	}

	tests := []TestCase{
		{
			name:      "empty body",
			content:   []byte{},
			signature: "",
		},
		{
			name:      "not empty body",
			content:   defaultContent,
			signature: defaultSignature,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ww := httpsec.NewResponseWriterWithHashHeader(w, headerKey, hasher.Sign)

			ww.Header().Set(xHeader, xValue)
			ww.WriteHeader(http.StatusTeapot)
			_, err := ww.Write(tc.content)

			require.NoError(t, err)
			require.NoError(t, ww.Finalize())

			resp := w.Result()
			content, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.NoError(t, resp.Body.Close())

			assert.Equal(t, http.StatusTeapot, resp.StatusCode)
			assert.Equal(t, xValue, resp.Header.Get(xHeader))
			assert.Equal(t, tc.signature, resp.Header.Get(headerKey))
			assert.Equal(t, tc.content, content)
		})
	}
}
