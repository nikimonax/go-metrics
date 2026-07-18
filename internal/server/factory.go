package server

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nikimonax/go-metrics/internal/app"
	"github.com/nikimonax/go-metrics/internal/impl"
	"github.com/nikimonax/go-metrics/internal/lib/zapextra"

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

	plainTextErrorPresenter := NewPlainTextErrorPresenter()
	plainTextMetricPresenter := NewPlainTextMetricPresenter()
	htmlTableMetricsPresenter := NewHtmlTableMetricsPresenter()

	updateMetricHandler := NewUpdateMetricHandler(
		updateMetricUseCase,
		plainTextErrorPresenter,
	)
	getMetricHandler := NewGetMetricHandler(
		getMetricUseCase,
		plainTextErrorPresenter,
		plainTextMetricPresenter,
	)
	PreviewMetricsHandler := NewPreviewMetricsHandler(
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
	router.Post(
		"/update/{metricType}/{metricName}/{metricValue}",
		updateMetricHandler.ServeHTTP,
	)
	router.Get(
		"/value/{metricType}/{metricName}",
		getMetricHandler.ServeHTTP,
	)

	return &Server{
		config: config,
		logger: logger,
		router: router,
	}
}
