package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/go-metrics/internal/server/config"
	"github.com/nikimonax/go-metrics/internal/server/middleware"
)

func TestAllowIPs(t *testing.T) {
	newHandler := func(t *testing.T, allowed ...string) http.Handler {
		t.Helper()

		allowIPs, err := middleware.AllowIPs(allowed...)
		require.NoError(t, err)

		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		return middleware.ClientIPFromHeader(config.SourceIPHeader)(
			allowIPs(next),
		)
	}

	doRequest := func(t *testing.T, handler http.Handler, headerValue string) *httptest.ResponseRecorder {
		t.Helper()

		req := httptest.NewRequestWithContext(
			t.Context(),
			http.MethodGet,
			"/test",
			nil,
		)

		if headerValue != "" {
			req.Header.Set(config.SourceIPHeader, headerValue)
		}

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		return rec
	}

	tests := []struct {
		name        string
		allowed     []string
		headerValue string
		wantStatus  int
	}{
		{
			name:        "ipv4 inside allowed subnet",
			allowed:     []string{"192.168.0.0/24"},
			headerValue: "192.168.0.15",
			wantStatus:  http.StatusOK,
		},
		{
			name:        "ipv4 outside allowed subnet",
			allowed:     []string{"192.168.0.0/24"},
			headerValue: "10.0.0.15",
			wantStatus:  http.StatusForbidden,
		},
		{
			name:        "bare ip matched",
			allowed:     []string{"10.1.2.3"},
			headerValue: "10.1.2.3",
			wantStatus:  http.StatusOK,
		},
		{
			name:        "bare ip not matched",
			allowed:     []string{"10.1.2.3"},
			headerValue: "10.1.2.4",
			wantStatus:  http.StatusForbidden,
		},
		{
			name:        "ipv4 mapped ipv6 is unmapped",
			allowed:     []string{"192.168.0.0/24"},
			headerValue: "::ffff:192.168.0.7",
			wantStatus:  http.StatusOK,
		},
		{
			name:        "ipv6 inside allowed subnet",
			allowed:     []string{"fd00::/8"},
			headerValue: "fd12:3456::1",
			wantStatus:  http.StatusOK,
		},
		{
			name:        "ipv6 outside allowed subnet",
			allowed:     []string{"fd00::/8"},
			headerValue: "2001:db8::1",
			wantStatus:  http.StatusForbidden,
		},
		{
			name:        "multiple entries: any match allows",
			allowed:     []string{"10.0.0.0/8", "192.168.0.0/24"},
			headerValue: "192.168.0.1",
			wantStatus:  http.StatusOK,
		},
		{
			name:        "multiple entries: no match forbids",
			allowed:     []string{"10.0.0.0/8", "192.168.0.0/24"},
			headerValue: "203.0.113.1",
			wantStatus:  http.StatusForbidden,
		},
		{
			name:        "no allowed entries forbids everyone",
			allowed:     nil,
			headerValue: "192.168.0.1",
			wantStatus:  http.StatusForbidden,
		},
		{
			name:        "missing header forbids",
			allowed:     []string{"192.168.0.0/24"},
			headerValue: "",
			wantStatus:  http.StatusForbidden,
		},
		{
			name:        "malformed header forbids",
			allowed:     []string{"192.168.0.0/24"},
			headerValue: "not-an-ip",
			wantStatus:  http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := newHandler(t, tc.allowed...)
			rec := doRequest(t, handler, tc.headerValue)

			assert.Equal(t, tc.wantStatus, rec.Code)
		})
	}

	t.Run("invalid entries return error", func(t *testing.T) {
		mw, err := middleware.AllowIPs("not-an-ip", "192.168.0.0/24", "bad..entry")

		require.Error(t, err)
		assert.Nil(t, mw)
		assert.ErrorContains(t, err, "invalid IP or CIDR: not-an-ip")
		assert.ErrorContains(t, err, "invalid IP or CIDR: bad..entry")
	})
}
