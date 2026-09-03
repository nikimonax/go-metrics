package shared

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

const DefaultURL = "http://example.com"

func NewRequest(t *testing.T, method string, body io.Reader) *http.Request {
	t.Helper()
	return httptest.NewRequestWithContext(t.Context(), method, DefaultURL, body)
}
