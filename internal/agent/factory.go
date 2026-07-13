package agent

import (
	"fmt"
	"time"

	"github.com/nikimonax/go-metrics/internal/app"
	"github.com/nikimonax/go-metrics/internal/impl"
)

type Agent struct {
	config                *AgentConfig
	collectMetricsUseCase *app.CollectMetricsUseCase
	sendMetricsUseCase    *app.SendMetricsUseCase
}

func (a *Agent) Run() {
	fmt.Printf(
		"Starting agent\n"+
			"  server: %s\n"+
			"  poll interval: %d secs\n"+
			"  send interval: %d secs\n",
		a.config.BaseURL,
		a.config.PollInterval/time.Second,
		a.config.ReportInterval/time.Second,
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

	NewScheduler(time.Now).Run(tasks)
}

func New(config *AgentConfig) *Agent {
	metricCollector := impl.NewCollectorsGroup(
		impl.CollectorFunc(impl.CollectMemStats),
		impl.CollectorFunc(impl.CollectRandomValue),
		impl.CollectorFunc(impl.CollectIncrOne),
	)

	metricGateway := impl.NewHttpMetricGateway(
		config.BaseURL,
	)

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
		collectMetricsUseCase: collectMetricsUseCase,
		sendMetricsUseCase:    sendMetricsUseCase,
	}
}
