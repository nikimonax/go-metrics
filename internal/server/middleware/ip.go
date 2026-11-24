package middleware

import (
	"net/http"
	"net/netip"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
)

func AllowIPs(prefixes []netip.Prefix) httpextra.Middleware {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			addr := middleware.GetClientIPAddr(r.Context())

			if !addr.IsValid() {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			addr = addr.Unmap()

			for _, prefix := range prefixes {
				if prefix.Contains(addr) {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, "Forbidden", http.StatusForbidden)
		}

		return http.HandlerFunc(fn)
	}
}
