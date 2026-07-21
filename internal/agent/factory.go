package agent

import (
	"log"
	"time"

	"github.com/nikimonax/go-metrics/internal/app"
	"github.com/nikimonax/go-metrics/internal/impl"
	"github.com/nikimonax/go-metrics/internal/lib/zapextra"

	"go.uber.org/zap"
)

type Agent struct {
	config                *AgentConfig
	logger                *zap.Logger
	collectMetricsUseCase *app.CollectMetricsUseCase
	sendMetricsUseCase    *app.SendMetricsUseCase
}

func (a *Agent) Run() {
	a.logger.Sugar().Infow(
		"starting agent",
		"server", a.config.BaseURL,
		"poll interval", a.config.PollInterval,
		"send interval", a.config.ReportInterval,
	)

	tasks := []Task{
		{
			Name:     "collect metrics",
			Interval: a.config.PollInterval,
			Callback: a.collectMetricsUseCase.Execute,
		},
		{
			Name:     "send metrics",
			Interval: a.config.ReportInterval,
			Callback: a.sendMetricsUseCase.Execute,
		},
	}

	NewScheduler(time.Now, a.logger).Run(tasks)
}

func New(config *AgentConfig) *Agent {
	logger := zapextra.NewZapLogger(zapextra.EnvDev)

	metricCollector := impl.NewCollectorsGroup(
		impl.CollectorFunc(impl.CollectMemStats),
		impl.CollectorFunc(impl.CollectRandomValue),
		impl.CollectorFunc(impl.CollectIncrOne),
	)

	var metricGateway app.MetricGateway

	switch config.ApiVersion {
	case 1:
		metricGateway = impl.NewHttpMetricGateway(config.BaseURL)
	case 2:
		metricGateway = impl.NewHttpMetricV2Gateway(config.BaseURL)
	default:
		log.Fatalf("unknown metrics server api version: %d", config.ApiVersion)
	}

	metricRepository := impl.NewInMemoryMetricRepository()

	collectMetricsUseCase := app.NewCollectMetricsUseCase(
		metricCollector,
		metricRepository,
	)

	sendMetricsUseCase := app.NewSendMetricsUseCase(
		metricGateway,
		metricRepository,
	)

	return &Agent{
		config:                config,
		logger:                logger,
		collectMetricsUseCase: collectMetricsUseCase,
		sendMetricsUseCase:    sendMetricsUseCase,
	}
}
