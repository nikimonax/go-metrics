package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"net/netip"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
)

func AllowIPs(allowed ...string) (httpextra.Middleware, error) {
	var (
		prefixes []netip.Prefix
		errs     error
	)

	for _, entry := range allowed {
		if prefix, err := netip.ParsePrefix(entry); err == nil {
			prefixes = append(prefixes, prefix)
			continue
		}

		if addr, err := netip.ParseAddr(entry); err == nil {
			prefixes = append(prefixes, netip.PrefixFrom(addr, addr.BitLen()))
			continue
		}

		err := fmt.Errorf("invalid IP or CIDR: %s", entry)
		errs = errors.Join(errs, err)
	}

	if errs != nil {
		return nil, errs
	}

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			addr := middleware.GetClientIPAddr(r.Context())

			if !addr.IsValid() {
				http.Error(w, "Forbidden", http.StatusForbidden)
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
	}, nil
}
