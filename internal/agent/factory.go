package agent

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/app/usecase"
	"github.com/nikimonax/go-metrics/internal/impl/collector"
	"github.com/nikimonax/go-metrics/internal/impl/gateway"
	"github.com/nikimonax/go-metrics/internal/impl/repository"
	"github.com/nikimonax/go-metrics/internal/lib/lifespan"
	"github.com/nikimonax/go-metrics/internal/lib/work"
	"github.com/nikimonax/go-metrics/internal/lib/zapextra"

	"go.uber.org/zap"
)

type Agent struct {
	config   *AgentConfig
	logger   *zap.Logger
	lifespan *lifespan.Lifespan
}

func (a *Agent) Run() {
	sugar := a.logger.Sugar()

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer cancel()

	sugar.Infow(
		"starting agent",
		"server", a.config.BaseURL,
		"poll interval", a.config.PollInterval,
		"send interval", a.config.ReportInterval,
	)

	if err := a.lifespan.Open(ctx); err != nil {
		sugar.Errorw("failed open lifespan", "err", err)
		return
	}

	<-ctx.Done()

	if err := a.lifespan.Close(context.Background()); err != nil {
		sugar.Errorw("failed close lifespan", "err", err)
	}
}

func New(config *AgentConfig) *Agent {
	logger := zapextra.NewZapLogger(zapextra.EnvDev)
	sugar := logger.Sugar()

	var err error

	poolConfig := work.PoolConfig{
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
	pool, err := work.NewPool(poolConfig)

	if err != nil {
		sugar.Fatalw(
			"failed create worker pool",
			"err", err,
		)
	}

	schedulerConfig := work.SchedulerConfig{
		LifecycleConfig: work.LifecycleConfig{
			StopTimeout: time.Second * 5,
		},
		OnError: func(name string, err error) {
			sugar.Errorw("failed submit task", "task", name, "err", err)
		},
	}
	scheduler, err := work.NewScheduler(pool, schedulerConfig)

	if err != nil {
		sugar.Fatalw(
			"failed create scheduler",
			"err", err,
		)
	}

	metricCollector := collector.NewCollectorsGroup(
		collector.CollectorFunc(collector.CollectMemStats),
		collector.CollectorFunc(collector.CollectRandomValue),
		collector.CollectorFunc(collector.CollectIncrOne),
	)

	var metricGateway interfaces.MetricGateway

	switch config.APIVersion {
	case 1:
		metricGateway = gateway.NewHTTPMetricGateway(config.BaseURL)
	case 2:
		metricGateway = gateway.NewHTTPMetricV2Gateway(config.BaseURL)
	default:
		sugar.Fatalw("unknown metrics server api version", "version", config.APIVersion)
	}

	// TODO: в usecase, repository, gateway и т.п. расширить интерфейсы,
	// пробрасывать context первым аргументом

	metricRepository := repository.NewInMemoryMetricRepository()

	collectMetricsUseCase := usecase.NewCollectMetricsUseCase(
		metricCollector,
		metricRepository,
	)

	sendMetricsUseCase := usecase.NewSendMetricsUseCase(
		metricGateway,
		metricRepository,
	)

	collectMetricsTaskName := "collect metrics"
	collectMetricsTask, err := work.NewTask(
		collectMetricsTaskName,
		func(_ context.Context) error {
			return collectMetricsUseCase.Execute()
		},
	)

	if err != nil {
		sugar.Fatalw(
			"failed create task",
			"task", collectMetricsTaskName,
			"err", err,
		)
	}

	err = scheduler.Schedule(
		collectMetricsTask,
		config.PollInterval,
	)

	if err != nil {
		sugar.Fatalw(
			"failed schedule",
			"task", collectMetricsTaskName,
			"err", err,
		)
	}

	sendMetricsTaskName := "send metrics"
	sendMetricsTask, err := work.NewTask(
		sendMetricsTaskName,
		func(_ context.Context) error {
			return sendMetricsUseCase.Execute()
		},
	)

	if err != nil {
		sugar.Fatalw(
			"failed create task",
			"task", sendMetricsTaskName,
			"err", err,
		)
	}

	err = scheduler.Schedule(
		sendMetricsTask,
		config.ReportInterval,
	)

	if err != nil {
		sugar.Fatalw(
			"failed schedule",
			"task", sendMetricsTaskName,
			"err", err,
		)
	}

	lifespan := lifespan.New()
	lifespan.OnStartup(pool.Start)
	lifespan.OnStartup(scheduler.Start)
	lifespan.OnShutdown(scheduler.Stop)
	lifespan.OnShutdown(pool.Stop)

	return &Agent{
		config:   config,
		logger:   logger,
		lifespan: lifespan,
	}
}
