package server

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/nikimonax/go-metrics/internal/app"
	"github.com/nikimonax/go-metrics/internal/impl"
	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	"github.com/nikimonax/go-metrics/internal/lib/lifespan"
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
	lifespan *lifespan.Lifespan
	router   chi.Router
}

func (s *Server) Run() {
	sugar := s.logger.Sugar()

	sugar.Infow(
		"starting server",
		"listen", s.config.BaseURL,
		"dump_file", s.config.DumpFile,
		"dump_interval", s.config.DumpInterval,
		"dump_restore", s.config.DumpRestore,
	)

	appCtx, appCancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer appCancel()

	if err := s.lifespan.Open(appCtx); err != nil {
		sugar.Errorw("failed open lifespan", "err", err)
		return
	}

	defer func() {
		shutdownCtx := context.Background()

		if timeout := s.config.LifespanCloseTimeout; timeout > 0 {
			var shutdownCancel context.CancelFunc
			shutdownCtx, shutdownCancel = context.WithTimeout(shutdownCtx, timeout)
			defer shutdownCancel()
		}

		if err := s.lifespan.Close(shutdownCtx); err != nil {
			sugar.Errorw("failed close lifespan", "err", err)
		}
	}()

	httpServer := &http.Server{
		Addr:    s.config.BaseURL,
		Handler: s.router,
	}

	serverErr := make(chan error, 1)

	go func() {
		serverErr <- httpServer.ListenAndServe()
	}()

	select {
	case <-appCtx.Done():
		shutdownCtx := context.Background()

		if timeout := s.config.ServerStopTimeout; timeout > 0 {
			var shutdownCancel context.CancelFunc
			shutdownCtx, shutdownCancel = context.WithTimeout(shutdownCtx, timeout)
			defer shutdownCancel()
		}

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			sugar.Errorw("failed server shutdown", "err", err)
		}

		if err := <-serverErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
			sugar.Errorw("failed listen and serve", "err", err)
		}

	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			sugar.Errorw("failed listen and serve", "err", err)
		}
	}
}

func New(config *ServerConfig) *Server {
	logger := zapextra.NewZapLogger(zapextra.EnvDev)
	sugar := logger.Sugar()

	lifespan := lifespan.New()

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
		lifespan.OnStartup(
			func(_ context.Context) error {
				return restoreMetricsUseCase.Execute()
			},
		)
	}

	updateMetricsHook := mymiddleware.NewRequestHook()

	if config.DumpInterval == 0 {
		saveMetricsUseCase := app.NewSaveMetricsUseCase(dumper, metricRepository)
		updateMetricsHook.AfterRequest(
			func(r *http.Request) {
				saveMetricsUseCase.Execute()
			},
		)
	}

	if config.DumpInterval > 0 {
		saveMetricsUseCase := app.NewSaveMetricsUseCase(dumper, metricRepository)

		scheduler := scheduler.New()
		scheduler.OnError = func(name string, err error) {
			sugar.Errorw("task failed", "task", name, "err", err)
		}
		scheduler.Schedule(
			"dump metrics",
			config.DumpInterval,
			func(_ context.Context) error {
				return saveMetricsUseCase.Execute()
			},
		)

		lifespan.OnStartup(scheduler.Start)
		lifespan.OnShutdown(scheduler.Stop)
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
