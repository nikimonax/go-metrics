package agent

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/nikimonax/go-metrics/internal/agent/config"
	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/app/usecase"
	"github.com/nikimonax/go-metrics/internal/impl/collector"
	"github.com/nikimonax/go-metrics/internal/impl/gateway"
	"github.com/nikimonax/go-metrics/internal/impl/repository"
	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	"github.com/nikimonax/go-metrics/internal/lib/httpsec"
	"github.com/nikimonax/go-metrics/internal/lib/work"
	"github.com/nikimonax/go-metrics/internal/lib/zapextra"

	"go.uber.org/fx"
	"go.uber.org/zap"
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
			provideRoundTripper,
			provideHTTPClient,
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
		fx.WithLogger(zapextra.NewFxLogger),
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
		collector.NewRuntimeStatsCollector(),
		collector.NewMemStatsCollector(),
		collector.NewCPUStatsCollector(),
		collector.NewRandomGaugeCollector("RandomValue"),
		collector.NewCounterCollector("PollCount", 1),
	)
}

func provideRoundTripper(cfg *config.AgentConfig) http.RoundTripper {
	transport := http.DefaultTransport

	if cfg.APIVersion > 1 && cfg.Security.HashingKey != "" {
		hasher := httpsec.NewHasher(
			cfg.Security.Header,
			cfg.Security.HashingFunc,
			[]byte(cfg.Security.HashingKey),
		)

		transport = hasher.RoundTripper(transport)
	}

	if cfg.APIVersion > 1 {
		transport = httpextra.NewCompressRoundTripper(transport, "gzip")
	}

	if cfg.RateLimit > 0 {
		transport = httpextra.NewRateLimitRoundTripper(transport, cfg.RateLimit)
	}

	return transport
}

func provideHTTPClient(transport http.RoundTripper) *http.Client {
	return &http.Client{Transport: transport}
}

func provideMetricGateway(
	cfg *config.AgentConfig,
	client *http.Client,
) (interfaces.MetricGateway, error) {
	gwFactory, err := gateway.GetGatewayFactory(cfg.APIVersion)

	if err != nil {
		return nil, err
	}

	gw := gwFactory(client, cfg.BaseURL)

	if cfg.Backoff.Retry == 0 {
		// without backoff
		return gw, nil
	}

	return gateway.NewRetryGateway(
		gw,
		cfg.Backoff.Build,
		gateway.HTTPErrorIsRetryable,
	), nil

}

func registerCollectMetricsTask(
	scheduler *work.Scheduler,
	cfg *config.AgentConfig,
	useCase *usecase.CollectMetricsUseCase,
) error {
	task, err := work.NewTask(
		"collect metrics",
		func(ctx context.Context) error {
			return useCase.Execute(ctx)
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
		func(ctx context.Context) error {
			return useCase.Execute(ctx)
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
	var hasSecretKey bool
	if cfg.Security.HashingKey != "" {
		hasSecretKey = true
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			sugar.Infow(
				"starting agent",
				"server", cfg.BaseURL,
				"api", "v"+fmt.Sprint(cfg.APIVersion),
				"rate_limit", cfg.RateLimit,
				"poll interval", cfg.PollInterval,
				"send interval", cfg.ReportInterval,
				"has_secret", hasSecretKey,
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
