package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/nikimonax/go-metrics/internal/agent/config"
	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/app/usecase"
	"github.com/nikimonax/go-metrics/internal/impl/collector"
	"github.com/nikimonax/go-metrics/internal/impl/gateway"
	"github.com/nikimonax/go-metrics/internal/impl/repository"
	"github.com/nikimonax/go-metrics/internal/lib/work"
	"github.com/nikimonax/go-metrics/internal/lib/zapextra"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Agent struct {
	app *fx.App
}

func (agent *Agent) Run() {
	agent.app.Run()
}

func New(cfg *config.AgentConfig) (*Agent, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	app := fx.New(
		fx.Supply(cfg),
		fx.Provide(
			provideLogger,
			provideSugaredLogger,
			providePoolConfig,
			providePoolSubmitter,
			provideSchedulerConfig,
			provideMetricCollector,
			provideMetricGateway,
			repository.NewInMemoryMetricRepository,
			usecase.NewCollectMetricsUseCase,
			usecase.NewSendMetricsUseCase,
			work.NewPool,
			work.NewScheduler,
		),
		fx.Invoke(
			registerCollectMetricsTask,
			registerSendMetricsTask,
			registerLifecycleHooks,
		),
		fx.WithLogger(provideFxLogger),
	)

	if err := app.Err(); err != nil {
		return nil, err
	}

	return &Agent{app: app}, nil
}

func provideLogger() *zap.Logger {
	return zapextra.NewZapLogger(zapextra.EnvDev, zap.InfoLevel)
}

func provideSugaredLogger(logger *zap.Logger) *zap.SugaredLogger {
	return logger.Sugar()
}

func provideFxLogger(logger *zap.Logger) fxevent.Logger {
	fxLogger := &fxevent.ZapLogger{Logger: logger}
	fxLogger.UseLogLevel(zapcore.DebugLevel)
	return fxLogger
}

func providePoolConfig(sugar *zap.SugaredLogger) work.PoolConfig {
	return work.PoolConfig{
		WorkerConfig: work.WorkerConfig{
			OnError: func(name string, err error) {
				sugar.Errorw("task failed", "task", name, "err", err)
			},
			OnPanic: func(name string, v any) {
				sugar.Errorw("task panic", "task", name, "value", v)
			},
		},
		LifecycleConfig: work.LifecycleConfig{
			StopTimeout: time.Second * 5,
		},
		WorkerCount: 2,
		QueueSize:   2,
	}
}

func providePoolSubmitter(pool *work.WorkerPool) work.Submitter {
	return pool
}

func provideSchedulerConfig(sugar *zap.SugaredLogger) work.SchedulerConfig {
	return work.SchedulerConfig{
		LifecycleConfig: work.LifecycleConfig{
			StopTimeout: time.Second * 5,
		},
		OnError: func(name string, err error) {
			sugar.Errorw("failed submit task", "task", name, "err", err)
		},
	}
}

func provideMetricCollector() interfaces.MetricCollector {
	return collector.NewCollectorsGroup(
		collector.CollectorFunc(collector.CollectMemStats),
		collector.CollectorFunc(collector.CollectRandomValue),
		collector.CollectorFunc(collector.CollectIncrOne),
	)
}

func provideMetricGateway(cfg *config.AgentConfig) (interfaces.MetricGateway, error) {
	switch cfg.APIVersion {
	case 1:
		return gateway.NewHTTPMetricGateway(cfg.BaseURL), nil
	case 2:
		return gateway.NewHTTPMetricV2Gateway(cfg.BaseURL), nil
	default:
		err := fmt.Errorf("unknown metrics server api version: %d", cfg.APIVersion)
		return nil, err
	}
}

func registerCollectMetricsTask(
	scheduler *work.Scheduler,
	cfg *config.AgentConfig,
	useCase *usecase.CollectMetricsUseCase,
) error {
	task, err := work.NewTask(
		"collect metrics",
		func(_ context.Context) error {
			return useCase.Execute()
		},
	)

	if err != nil {
		return err
	}

	return scheduler.Schedule(task, cfg.PollInterval)
}

func registerSendMetricsTask(
	scheduler *work.Scheduler,
	cfg *config.AgentConfig,
	useCase *usecase.SendMetricsUseCase,
) error {
	task, err := work.NewTask(
		"send metrics",
		func(_ context.Context) error {
			return useCase.Execute()
		},
	)

	if err != nil {
		return err
	}

	return scheduler.Schedule(task, cfg.ReportInterval)
}

func registerLifecycleHooks(
	lc fx.Lifecycle,
	cfg *config.AgentConfig,
	sugar *zap.SugaredLogger,
	pool *work.WorkerPool,
	scheduler *work.Scheduler,
) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			sugar.Infow(
				"starting agent",
				"server", cfg.BaseURL,
				"poll interval", cfg.PollInterval,
				"send interval", cfg.ReportInterval,
			)
			return nil
		},
	})
	lc.Append(fx.Hook{
		OnStart: pool.Start,
		OnStop:  pool.Stop,
	})
	lc.Append(fx.Hook{
		OnStart: scheduler.Start,
		OnStop:  scheduler.Stop,
	})
}
