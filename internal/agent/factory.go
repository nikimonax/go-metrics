package agent

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/app/usecase"
	"github.com/nikimonax/go-metrics/internal/impl/collector"
	"github.com/nikimonax/go-metrics/internal/impl/gateway"
	"github.com/nikimonax/go-metrics/internal/impl/repository"
	"github.com/nikimonax/go-metrics/internal/lib/scheduler"
	"github.com/nikimonax/go-metrics/internal/lib/zapextra"

	"go.uber.org/zap"
)

type Agent struct {
	config    *AgentConfig
	logger    *zap.Logger
	scheduler *scheduler.Scheduler
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

	if err := a.scheduler.Run(ctx); err != nil {
		sugar.Errorw("scheduler stopped", "err", err)
	}
}

func New(config *AgentConfig) *Agent {
	logger := zapextra.NewZapLogger(zapextra.EnvDev)
	sugar := logger.Sugar()

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

	metricRepository := repository.NewInMemoryMetricRepository()

	collectMetricsUseCase := usecase.NewCollectMetricsUseCase(
		metricCollector,
		metricRepository,
	)

	sendMetricsUseCase := usecase.NewSendMetricsUseCase(
		metricGateway,
		metricRepository,
	)

	scheduler := scheduler.New()
	scheduler.OnError = func(name string, err error) {
		sugar.Errorw("task failed", "task", name, "err", err)
	}

	// TODO: в usecase, repository, gateway и т.п. расширить интерфейсы,
	// пробрасывать context первым аргументом

	var err error

	collectMetricsTaskName := "collect metrics"
	err = scheduler.Schedule(
		collectMetricsTaskName,
		config.PollInterval,
		func(_ context.Context) error {
			return collectMetricsUseCase.Execute()
		},
	)

	if err != nil {
		sugar.Fatalw(
			"failed schedule",
			"task", collectMetricsTaskName,
			"err", err,
		)
	}

	sendMetricsTaskName := "send metrics"
	err = scheduler.Schedule(
		sendMetricsTaskName,
		config.ReportInterval,
		func(_ context.Context) error {
			return sendMetricsUseCase.Execute()
		},
	)

	if err != nil {
		sugar.Fatalw(
			"failed schedule",
			"task", sendMetricsTaskName,
			"err", err,
		)
	}

	return &Agent{
		config:    config,
		logger:    logger,
		scheduler: scheduler,
	}
}
