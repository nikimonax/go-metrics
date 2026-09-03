package fxmodule

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	mymiddleware "github.com/nikimonax/go-metrics/internal/server/middleware"
)

type Middleware = func(http.Handler) http.Handler

func provideJSONRouter(
	baseRouter chi.Router,
	logger Middleware,
	compress Middleware,
) chi.Router {
	return baseRouter.With(
		middleware.AllowContentType(httpextra.MIMEJSON),
		middleware.AllowContentEncoding(httpextra.ENCGzip),
		mymiddleware.Decompress(),
		compress,
		logger,
	)
}
