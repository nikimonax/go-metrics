package fxmodule

import (
	"context"
	"errors"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/app/usecase"
	"github.com/nikimonax/go-metrics/internal/impl/publisher"
	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	"github.com/nikimonax/go-metrics/internal/lib/httpsec"
	"github.com/nikimonax/go-metrics/internal/lib/work"
	"github.com/nikimonax/go-metrics/internal/lib/zapextra"
	"github.com/nikimonax/go-metrics/internal/server/config"
	"github.com/nikimonax/go-metrics/internal/server/handler"
	"github.com/nikimonax/go-metrics/internal/server/middleware"
	"github.com/nikimonax/go-metrics/internal/server/presenter"
)

func CoreModule() fx.Option {
	return fx.Module(
		"core",
		// logging
		fx.Provide(
			provideLogger,
			provideSugaredLogger,
		),
		// utility
		fx.Provide(
			provideValidator,
			provideTranslator,
			provideHasher,
		),
		// background processing
		fx.Provide(
			providePoolConfig,
			providePoolSubmitter,
			providerSchedulerConfig,
			work.NewPool,
			work.NewScheduler,
		),
		// application level stuff
		fx.Provide(
			provideEventDispatcher,
			provideEventPublisher,
			usecase.NewUpdateMetricUseCase,
			usecase.NewUpdateMetricsUseCase,
			usecase.NewGetMetricUseCase,
			usecase.NewGetAllMetricsUseCase,
			provideHandlerUpdateMetricUseCase,
			provideHandlerUpdateMetricsUseCase,
			provideHandlerGetMetricUseCase,
			provideHandlerGetAllMetricsUseCase,
		),
		// presenters
		fx.Provide(
			fx.Annotate(
				presenter.NewPlainTextErrorPresenter,
				fx.ResultTags(`name:"presenter_error_text"`),
			),
			fx.Annotate(
				presenter.NewPlainTextMetricPresenter,
				fx.ResultTags(`name:"presenter_metric_text"`),
			),
			fx.Annotate(
				presenter.NewHTMLTableMetricsPresenter,
				fx.ResultTags(`name:"presenter_metrics_html"`),
			),
			fx.Annotate(
				presenter.NewJSONErrorPresenter,
				fx.ResultTags(`name:"presenter_error_json"`),
			),
			fx.Annotate(
				presenter.NewJSONMetricPresenter,
				fx.ResultTags(`name:"presenter_metric_json"`),
			),
		),
		// http routes, handlers, middleware
		fx.Provide(
			fx.Annotate(
				zapextra.NewZapSugarLoggingMiddleware,
				fx.ResultTags(`name:"middleware_logger"`),
			),
			fx.Annotate(
				providerMiddlewareCompress,
				fx.ResultTags(`name:"middleware_compress"`),
			),
			fx.Annotate(
				providerBaseRouter,
				fx.ResultTags(`name:"router_base"`),
			),
		),
		fx.Invoke(
			registerStartupLog,
			registerPoolLifecycle,
			registerSchedulerLifecycle,
			fx.Annotate(
				registerHTTPServerLifecycle,
				fx.ParamTags("", "", `name:"router_base"`),
			),
		),
	)
}

func provideLogger(cfg *config.ServerConfig) *zap.Logger {
	return zapextra.NewZapLogger(cfg.Log.Env, cfg.Log.Level)
}

func provideSugaredLogger(logger *zap.Logger) *zap.SugaredLogger {
	return logger.Sugar()
}

func provideValidator() *validator.Validate {
	return validator.New(validator.WithRequiredStructEnabled())
}

func provideTranslator() ut.Translator {
	return presenter.NewTranslator()
}

func provideHasher(cfg *config.ServerConfig) *httpsec.Hasher {
	if cfg.Security.HashingKey == "" {
		return nil
	}

	return httpsec.NewHasher(
		cfg.Security.Header,
		cfg.Security.HashingFunc,
		[]byte(cfg.Security.HashingKey),
	)
}

func providePoolConfig(
	cfg *config.ServerConfig,
	sugar *zap.SugaredLogger,
) work.PoolConfig {
	return work.PoolConfig{
		WorkerConfig: work.WorkerConfig{
			OnError: func(name string, err error) {
				sugar.Errorw("task failed", "task", name, "err", err)
			},
			OnPanic: func(name string, v any) {
				sugar.Errorw("task panic", "task", name, "value", v)
			},
		},
		LifecycleConfig: cfg.Pool.LifecycleConfig,
		WorkerCount:     cfg.Pool.WorkerCount,
		QueueSize:       cfg.Pool.QueueSize,
	}
}

func providerSchedulerConfig(
	cfg *config.ServerConfig,
	sugar *zap.SugaredLogger,
) work.SchedulerConfig {
	return work.SchedulerConfig{
		LifecycleConfig: cfg.Scheduler.LifecycleConfig,
		OnError: func(name string, err error) {
			sugar.Errorw("failed submit task", "task", name, "err", err)
		},
	}
}

func provideEventDispatcher(
	sugar *zap.SugaredLogger,
) *publisher.EventDispatcher {
	dp := publisher.NewEventDispatcher()
	dp.OnError = func(e interfaces.Event, err error) {
		sugar.Errorw("failed handle event", "event", e, "err", err)
	}
	return dp
}

func providePoolSubmitter(pool *work.WorkerPool) work.Submitter {
	return pool
}

func provideEventPublisher(dp *publisher.EventDispatcher) interfaces.EventPublisher {
	return dp
}

func provideHandlerUpdateMetricUseCase(
	usecase *usecase.UpdateMetricUseCase,
) handler.UpdateMetricUseCase {
	return usecase
}

func provideHandlerUpdateMetricsUseCase(
	usecase *usecase.UpdateMetricsUseCase,
) handler.UpdateMetricsUseCase {
	return usecase
}

func provideHandlerGetMetricUseCase(
	usecase *usecase.GetMetricUseCase,
) handler.GetMetricUseCase {
	return usecase
}

func provideHandlerGetAllMetricsUseCase(
	usecase *usecase.GetAllMetricsUseCase,
) handler.GetAllMetricsUseCase {
	return usecase
}

func providerMiddlewareCompress() httpextra.Middleware {
	return middleware.Compress(5)
}

func providerBaseRouter(hasher *httpsec.Hasher) chi.Router {
	baseRouter := chi.NewRouter()
	baseRouter.Use(middleware.CleanPath)

	if hasher != nil {
		baseRouter.Use(hasher.CalculateMiddleware())
	}

	return baseRouter
}

func registerStartupLog(
	lc fx.Lifecycle,
	sugar *zap.SugaredLogger,
	cfg *config.ServerConfig,
) {
	var hasSecretKey bool
	if cfg.Security.HashingKey != "" {
		hasSecretKey = true
	}

	lc.Append(fx.StartHook(func() {
		sugar.Infow(
			"starting server",
			"listen", cfg.Listen,
			"database", cfg.Database.DSN,
			"has_secret", hasSecretKey,
			"dump_file", cfg.Dump.File,
			"dump_interval", cfg.Dump.Interval,
			"dump_restore", cfg.Dump.Restore,
		)
	}))
}

func registerPoolLifecycle(
	lc fx.Lifecycle,
	pool *work.WorkerPool,
) {
	lc.Append(fx.Hook{
		OnStart: pool.Start,
		OnStop:  pool.Stop,
	})
}

func registerSchedulerLifecycle(
	lc fx.Lifecycle,
	scheduler *work.Scheduler,
) {
	lc.Append(fx.Hook{
		OnStart: scheduler.Start,
		OnStop:  scheduler.Stop,
	})
}

func registerHTTPServerLifecycle(
	lc fx.Lifecycle,
	cfg *config.ServerConfig,
	router chi.Router,
) {
	srv := &http.Server{
		Addr:    cfg.Listen,
		Handler: router,
	}

	errs := make(chan error, 1)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			lc := net.ListenConfig{}

			ln, err := lc.Listen(ctx, "tcp", srv.Addr)
			if err != nil {
				return err
			}

			go func() {
				err := srv.Serve(ln)

				if err != nil && err != http.ErrServerClosed {
					errs <- err
				}

				close(errs)
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			shutdownErr := srv.Shutdown(ctx)

			select {
			case <-ctx.Done():
				return shutdownErr
			case serveErr := <-errs:
				return errors.Join(shutdownErr, serveErr)
			}
		},
	})
}
