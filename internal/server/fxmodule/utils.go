package fxmodule

import (
	"github.com/go-chi/chi/v5"

	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	"github.com/nikimonax/go-metrics/internal/server/middleware"
)

func provideJSONRouter(
	baseRouter chi.Router,
	logger httpextra.Middleware,
	compress httpextra.Middleware,
) chi.Router {
	return baseRouter.With(
		middleware.AllowContentType(httpextra.MIMEJSON),
		middleware.AllowContentEncoding(httpextra.ENCGzip),
		middleware.Decompress(),
		compress,
		logger,
	)
}
