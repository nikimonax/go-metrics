package fxmodule

import (
	"context"

	"go.uber.org/fx"

	"github.com/nikimonax/go-metrics/internal/app/event"
	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/app/usecase"
	"github.com/nikimonax/go-metrics/internal/impl/dumper"
	"github.com/nikimonax/go-metrics/internal/impl/publisher"
	"github.com/nikimonax/go-metrics/internal/impl/serializer"
	"github.com/nikimonax/go-metrics/internal/lib/work"
	"github.com/nikimonax/go-metrics/internal/server/config"
)

func DumpModule() fx.Option {
	return fx.Module(
		"dump",
		fx.Provide(
			serializer.NewJSONMetricSerializer,
			provideMetricDumper,
		),
		fx.Provide(
			usecase.NewRestoreMetricsUseCase,
			usecase.NewSaveMetricsUseCase,
		),
		fx.Provide(
			fx.Annotate(
				provideSaveMetricsTask,
				fx.ResultTags(`name:"task_metrics_save"`),
			),
			fx.Annotate(
				publisher.NewSubmitTaskOnEventHandler,
				fx.ParamTags(`name:"task_metrics_save"`, ""),
				fx.ResultTags(`name:"event_handler_submit_task_metrics_save"`),
			),
		),
		fx.Invoke(
			registerRestore,
			fx.Annotate(
				registerSyncDumps,
				fx.ParamTags(
					"", "",
					`name:"event_handler_submit_task_metrics_save"`,
				),
			),
			fx.Annotate(
				registerPeriodicDumps,
				fx.ParamTags(
					"", "",
					`name:"task_metrics_save"`,
				),
			),
		),
	)
}

func provideMetricDumper(
	cfg *config.ServerConfig,
	serializer serializer.MetricSerializer,
) interfaces.MetricDumper {
	return dumper.NewFileMetricDumper(cfg.Dump.File, serializer)
}

func provideSaveMetricsTask(
	useCase *usecase.SaveMetricsUseCase,
) (work.Task, error) {
	return work.NewTask(
		"dump metrics",
		func(_ context.Context) error {
			// TODO: pass context
			return useCase.Execute()
		},
	)
}

func registerRestore(
	lc fx.Lifecycle,
	cfg *config.ServerConfig,
	restoreMetricsUseCase *usecase.RestoreMetricsUseCase,
) {
	if cfg.Dump.Restore {
		lc.Append(fx.StartHook(restoreMetricsUseCase.Execute))
	}

}

func registerSyncDumps(
	cfg *config.ServerConfig,
	dp *publisher.EventDispatcher,
	handler publisher.EventHandler,
) {
	if cfg.Dump.Interval != 0 {
		return
	}

	dp.Register(event.MetricsUpdatedEventName, handler)
}

func registerPeriodicDumps(
	cfg *config.ServerConfig,
	scheduler *work.Scheduler,
	task work.Task,
) error {
	if cfg.Dump.Interval <= 0 {
		return nil
	}

	return scheduler.Schedule(task, cfg.Dump.Interval)
}
