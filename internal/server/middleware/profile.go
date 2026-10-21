package middleware

import (
	"context"
	"net/http"
	"runtime/pprof"

	"github.com/go-chi/chi/v5"
)

func PprofLabels(router chi.Router) func(http.Handler) http.Handler {
	var mux *chi.Mux
	if m, ok := router.(*chi.Mux); ok {
		mux = m
	}

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			pprof.Do(
				r.Context(),
				pprof.Labels(
					"method", r.Method,
					"endpoint", routePattern(mux, r),
				),
				func(ctx context.Context) {
					next.ServeHTTP(w, r.WithContext(ctx))
				},
			)
		}
		return http.HandlerFunc(fn)
	}
}

func routePattern(mux *chi.Mux, r *http.Request) string {
	if mux != nil {
		if pattern := mux.Find(chi.NewRouteContext(), r.Method, r.URL.Path); pattern != "" {
			return pattern
		}
	}

	rctx := chi.RouteContext(r.Context())
	if rctx != nil {
		if pattern := rctx.RoutePattern(); pattern != "" {
			return pattern
		}
	}

	return r.URL.Path
}
