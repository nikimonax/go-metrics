package agent

import (
	"context"
	"fmt"
	"net/http"
	"os"

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
	"github.com/nikimonax/go-metrics/internal/shared/fxutil"
	"github.com/nikimonax/go-metrics/internal/shared/metainfo"
	"github.com/nikimonax/go-metrics/internal/shared/retry"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Agent struct {
	app *fx.App
}

func (agent *Agent) Run(ctx context.Context) error {
	return fxutil.RunWithContext(ctx, agent.app)
}

func New(cfg *config.AgentConfig) (*Agent, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	app := fx.New(
		fx.Supply(cfg),
		// logging
		fx.Provide(
			provideLogger,
			provideSugaredLogger,
		),
		// background processing
		fx.Provide(
			config.NewPoolConfig,
			config.NewSchedulerConfig,
			work.NewPool,
			work.NewScheduler,
			work.PoolAsSubmitter,
		),
		// usecases
		fx.Provide(
			provideRoundTripper,
			provideHTTPClient,
			provideMetricCollector,
			provideMetricGateway,
			fx.Annotate(
				repository.NewInMemoryMetricRepository,
				fx.As(new(interfaces.MetricRepository)),
			),
			usecase.NewCollectMetricsUseCase,
			usecase.NewSendMetricsUseCase,
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

func provideMetricCollector() interfaces.MetricCollector {
	return collector.NewCollectorsGroup(
		collector.NewRuntimeStatsCollector(),
		collector.NewMemStatsCollector(),
		collector.NewCPUStatsCollector(),
		collector.NewRandomGaugeCollector("RandomValue"),
		collector.NewCounterCollector("PollCount", 1),
	)
}

func provideRoundTripper(cfg *config.AgentConfig) (http.RoundTripper, error) {
	transport := http.DefaultTransport

	if cfg.APIVersion > 1 {
		newTransport, err := httpextra.NewCompressRoundTripper(
			transport,
			"gzip",
			config.CompressLevel,
		)

		if err != nil {
			return nil, err
		}

		transport = newTransport
	}

	if cfg.APIVersion > 1 && cfg.Security.HashingKey != "" {
		hasher := httpsec.NewHasher(
			cfg.Security.Header,
			cfg.Security.HashingFunc,
			[]byte(cfg.Security.HashingKey),
		)

		transport = hasher.RoundTripper(transport)
	}

	if cfg.APIVersion == 3 && cfg.Security.CryptoKeyFile != "" {
		key, err := cfg.Security.ReadPublicKey()

		if err != nil {
			return nil, err
		}

		encryptor := httpsec.NewEncryptor()
		transport, err = encryptor.EncryptRoundTripper(transport, key)

		if err != nil {
			return nil, err
		}
	}

	if cfg.RateLimit > 0 {
		transport = httpextra.NewRateLimitRoundTripper(transport, cfg.RateLimit)
	}

	return transport, nil
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
		retry.HTTPErrorIsRetryable,
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

	var hasCryptoKey bool
	if cfg.Security.CryptoKeyFile != "" {
		hasCryptoKey = true
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			fmt.Fprintf(os.Stderr, "Build version: %s\n", metainfo.BuildVersion)
			fmt.Fprintf(os.Stderr, "Build date: %s\n", metainfo.BuildDate)
			fmt.Fprintf(os.Stderr, "Build commit: %s\n", metainfo.BuildCommit)
			sugar.Infow(
				"starting agent",
				"server", cfg.BaseURL,
				"api", "v"+fmt.Sprint(cfg.APIVersion),
				"rate_limit", cfg.RateLimit,
				"poll interval", cfg.PollInterval,
				"send interval", cfg.ReportInterval,
				"has_secret", hasSecretKey,
				"has_crypto", hasCryptoKey,
			)
			return nil
		},
		OnStop: func(_ context.Context) error {
			sugar.Infow("stopping agent")
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
