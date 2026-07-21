package server

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nikimonax/go-metrics/internal/app"
	"github.com/nikimonax/go-metrics/internal/impl"
	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	"github.com/nikimonax/go-metrics/internal/lib/zapextra"
	"github.com/nikimonax/go-metrics/internal/server/handler"
	"github.com/nikimonax/go-metrics/internal/server/presenter"

	"go.uber.org/zap"
)

type Server struct {
	config *ServerConfig
	logger *zap.Logger
	router chi.Router
}

func (s *Server) Run() {
	sugar := s.logger.Sugar()
	sugar.Infow(
		"starting server",
		"listen", s.config.BaseURL,
	)

	err := http.ListenAndServe(s.config.BaseURL, s.router)

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		sugar.Errorf("failed start server: %s", err)
	}
}

func New(config *ServerConfig) *Server {
	logger := zapextra.NewZapLogger(zapextra.EnvDev)

	metricRepository := impl.NewInMemoryMetricRepository()

	updateMetricUseCase := app.NewUpdateMetricUseCase(metricRepository)
	getMetricUseCase := app.NewGetMetricUseCase(metricRepository)
	getAllMetricsUseCase := app.NewGetAllMetricsUseCase(metricRepository)

	plainTextErrorPresenter := presenter.NewPlainTextErrorPresenter()
	jsonErrorPresenter := presenter.NewJsonErrorPresenter(logger)
	plainTextMetricPresenter := presenter.NewPlainTextMetricPresenter(logger)
	jsonMetricPresenter := presenter.NewJsonMetricPresenter(logger)
	htmlTableMetricsPresenter := presenter.NewHtmlTableMetricsPresenter(logger)

	updateMetricHandler := handler.NewUpdateMetricHandler(
		updateMetricUseCase,
		plainTextErrorPresenter,
	)
	updateMetricHandlerV2 := handler.NewUpdateMetricV2Handler(
		updateMetricUseCase,
		jsonErrorPresenter,
	)
	getMetricHandler := handler.NewGetMetricHandler(
		getMetricUseCase,
		plainTextErrorPresenter,
		plainTextMetricPresenter,
	)
	getMetricHandlerV2 := handler.NewGetMetricV2Handler(
		getMetricUseCase,
		jsonErrorPresenter,
		jsonMetricPresenter,
	)
	PreviewMetricsHandler := handler.NewPreviewMetricsHandler(
		getAllMetricsUseCase,
		plainTextErrorPresenter,
		htmlTableMetricsPresenter,
	)

	router := chi.NewRouter()
	router.Use(zapextra.NewZapSugarLoggingMiddleware(logger))

	router.Get(
		"/",
		PreviewMetricsHandler.ServeHTTP,
	)

	// api v1 (спринт 1 - path params)
	router.Post(
		"/update/{metricType}/{metricName}/{metricValue}",
		updateMetricHandler.ServeHTTP,
	)
	router.Get(
		"/value/{metricType}/{metricName}",
		getMetricHandler.ServeHTTP,
	)

	// api v2 (спринт 2 - json payload)
	router.With(
		middleware.AllowContentType(httpextra.MIMEJSON),
	).Post(
		"/update",
		updateMetricHandlerV2.ServeHTTP,
	)
	router.With(
		middleware.AllowContentType(httpextra.MIMEJSON),
	).Post(
		"/value",
		getMetricHandlerV2.ServeHTTP,
	)

	return &Server{
		config: config,
		logger: logger,
		router: router,
	}
}
