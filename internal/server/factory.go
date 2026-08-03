package server

import (
	"errors"
	"net/http"

	"github.com/nikimonax/go-metrics/internal/app"
	"github.com/nikimonax/go-metrics/internal/impl"
	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	"github.com/nikimonax/go-metrics/internal/lib/scheduler"
	"github.com/nikimonax/go-metrics/internal/lib/zapextra"
	"github.com/nikimonax/go-metrics/internal/server/handler"
	mymiddleware "github.com/nikimonax/go-metrics/internal/server/middleware"
	"github.com/nikimonax/go-metrics/internal/server/presenter"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type Server struct {
	config   *ServerConfig
	logger   *zap.Logger
	lifespan *Lifespan
	router   chi.Router
}

func (s *Server) Run() {
	sugar := s.logger.Sugar()
	sugar.Infow(
		"starting server",
		"listen", s.config.BaseURL,
	)

	err := s.lifespan.Open()

	if err == nil {
		defer s.lifespan.Close()
		err = http.ListenAndServe(s.config.BaseURL, s.router)
	}

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		sugar.Errorf("failed start server: %s", err)
	}
}

func New(config *ServerConfig) *Server {
	logger := zapextra.NewZapLogger(zapextra.EnvDev)
	sugar := logger.Sugar()

	lifespan := NewLifespan()

	metricRepository := impl.NewInMemoryMetricRepository()

	updateMetricUseCase := app.NewUpdateMetricUseCase(metricRepository)
	getMetricUseCase := app.NewGetMetricUseCase(metricRepository)
	getAllMetricsUseCase := app.NewGetAllMetricsUseCase(metricRepository)

	plainTextErrorPresenter := presenter.NewPlainTextErrorPresenter()
	jsonErrorPresenter := presenter.NewJsonErrorPresenter(logger)
	plainTextMetricPresenter := presenter.NewPlainTextMetricPresenter(logger)
	jsonMetricPresenter := presenter.NewJsonMetricPresenter(logger)
	htmlTableMetricsPresenter := presenter.NewHtmlTableMetricsPresenter(logger)

	var dumper app.MetricDumper
	if config.DumpFile != "" {
		serializer := impl.NewJsonMetricSerializer()
		dumper = impl.NewFileMetricDumper(config.DumpFile, serializer)
	}

	if config.DumpRestore {
		restoreMetricsUseCase := app.NewRestoreMetricsUseCase(dumper, metricRepository)
		lifespan.OnStartup(restoreMetricsUseCase.Execute)
	}

	updateMetricsHook := mymiddleware.NewRequestHook()

	if config.DumpInterval == 0 {
		saveMetricsUseCase := app.NewSaveMetricsUseCase(dumper, metricRepository)
		updateMetricsHook.AfterRequest(func(r *http.Request) { saveMetricsUseCase.Execute() })
	}

	if config.DumpInterval > 0 {
		saveMetricsUseCase := app.NewSaveMetricsUseCase(dumper, metricRepository)

		s := scheduler.New()
		s.OnError = func(name string, err error) {
			sugar.Errorw("task failed", "task", name, "err", err)
		}
		s.Schedule("dump metrics", config.DumpInterval, saveMetricsUseCase.Execute)

		lifespan.OnStartup(s.Start)
		lifespan.OnShutdown(s.Stop)
	}

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
	middlewareCompress := middleware.Compress(5)

	baseRouter := chi.NewRouter()
	baseRouter.Use(middleware.CleanPath)

	// api v1 (спринт 1 - path params)
	routerV1 := baseRouter.With(
		middlewareCompress,
		middlewareLogger,
	)

	routerV1.Get(
		"/",
		PreviewMetricsHandler.ServeHTTP,
	)

	routerV1.With(
		updateMetricsHook.Middleware,
	).Post(
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
		middleware.AllowContentEncoding(httpextra.ENCGzip),
		mymiddleware.Decompress(),
		middlewareCompress,
		middlewareLogger,
	)

	routerV2.With(
		updateMetricsHook.Middleware,
	).Post(
		"/update",
		updateMetricHandlerV2.ServeHTTP,
	)
	routerV2.Post(
		"/value",
		getMetricHandlerV2.ServeHTTP,
	)

	return &Server{
		config:   config,
		logger:   logger,
		lifespan: lifespan,
		router:   baseRouter,
	}
}
