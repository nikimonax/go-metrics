package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/go-metrics/internal/lib/netutil"
	"github.com/nikimonax/go-metrics/internal/server/config"
	"github.com/nikimonax/go-metrics/internal/server/middleware"
)

func TestAllowIPs(t *testing.T) {
	newHandler := func(t *testing.T, allowed ...string) http.Handler {
		t.Helper()

		prefixes, err := netutil.ParsePrefixes(allowed)
		require.NoError(t, err)

		return middleware.ClientIPFromHeader(config.SourceIPKey)(
			middleware.AllowIPs(prefixes)(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			),
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
			req.Header.Set(config.SourceIPKey, headerValue)
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
		wantContent string
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
			wantContent: "Forbidden\n",
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
			wantContent: "Forbidden\n",
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
			wantContent: "Forbidden\n",
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
			wantContent: "Forbidden\n",
		},
		{
			name:        "no allowed entries forbids everyone",
			allowed:     nil,
			headerValue: "192.168.0.1",
			wantStatus:  http.StatusForbidden,
			wantContent: "Forbidden\n",
		},
		{
			name:        "missing header forbids",
			allowed:     []string{"192.168.0.0/24"},
			headerValue: "",
			wantStatus:  http.StatusForbidden,
			wantContent: "Forbidden\n",
		},
		{
			name:        "malformed header forbids",
			allowed:     []string{"192.168.0.0/24"},
			headerValue: "not-an-ip",
			wantStatus:  http.StatusForbidden,
			wantContent: "Forbidden\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := newHandler(t, tc.allowed...)
			rec := doRequest(t, handler, tc.headerValue)

			assert.Equal(t, tc.wantStatus, rec.Code)

			if tc.wantContent != "" {
				assert.Equal(t, tc.wantContent, rec.Body.String())
			}
		})
	}
}
