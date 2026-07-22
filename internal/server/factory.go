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

	middlewareLogger := zapextra.NewZapSugarLoggingMiddleware(logger)

	baseRouter := chi.NewRouter()
	baseRouter.Use(middleware.CleanPath)

	// api v1 (спринт 1 - path params)
	routerV1 := baseRouter.With(
		middlewareLogger,
	)

	routerV1.Get(
		"/",
		PreviewMetricsHandler.ServeHTTP,
	)

	routerV1.Post(
		"/update/{metricType}/{metricName}/{metricValue}",
		updateMetricHandler.ServeHTTP,
	)
	routerV1.Get(
		"/value/{metricType}/{metricName}",
		getMetricHandler.ServeHTTP,
	)

	// api v2 (спринт 2 - json payload)
	routerV2 := baseRouter.With(
		middleware.AllowContentType(httpextra.MIMEJSON),
		middleware.Compress(5),
		middlewareLogger,
	)

	routerV2.Post(
		"/update",
		updateMetricHandlerV2.ServeHTTP,
	)
	routerV2.Post(
		"/value",
		getMetricHandlerV2.ServeHTTP,
	)

	return &Server{
		config: config,
		logger: logger,
		router: baseRouter,
	}
}
