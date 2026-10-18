package fxmodule

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"

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
		// usecases
		fx.Provide(
			serializer.NewJSONMetricSerializer,
			provideMetricDumper,
			usecase.NewRestoreMetricsUseCase,
			usecase.NewSaveMetricsUseCase,
			fx.Private,
		),
		// background processing
		fx.Provide(
			fx.Annotate(
				provideDumpPoolConfig,
				fx.ResultTags(`name:"pool_config_dump"`),
			),
			fx.Annotate(
				config.NewSchedulerConfig,
				fx.ResultTags(`name:"scheduler_config_dump"`),
			),
			fx.Annotate(
				work.NewPool,
				fx.ParamTags(`name:"pool_config_dump"`),
				fx.ResultTags(`name:"pool_dump"`),
			),
			fx.Annotate(
				work.NewScheduler,
				fx.ParamTags(`name:"submitter_dump"`, `name:"scheduler_config_dump"`),
				fx.ResultTags(`name:"scheduler_dump"`),
			),
			fx.Annotate(
				work.PoolAsSubmitter,
				fx.ParamTags(`name:"pool_dump"`),
				fx.ResultTags(`name:"submitter_dump"`),
			),
			fx.Annotate(
				provideSaveMetricsTask,
				fx.ResultTags(`name:"task_metrics_save"`),
			),
			fx.Annotate(
				publisher.NewSubmitTaskOnEventHandler,
				fx.ParamTags(`name:"task_metrics_save"`, `name:"submitter_dump"`),
				fx.ResultTags(`name:"event_handler_submit_task_metrics_save"`),
			),
		),
		fx.Invoke(
			fx.Annotate(
				registerPoolLifecycle,
				fx.ParamTags("", `name:"pool_dump"`),
			),
			fx.Annotate(
				registerSchedulerLifecycle,
				fx.ParamTags("", `name:"scheduler_dump"`),
			),
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
					"",
					`name:"scheduler_dump"`,
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

func provideDumpPoolConfig(
	sugar *zap.SugaredLogger,
) work.PoolConfig {
	return work.PoolConfig{
		WorkerConfig:    config.NewWorkerConfig(sugar),
		LifecycleConfig: config.NewLifecycleConfig(),
		WorkerCount:     1,
		QueueSize:       1,
	}
}

func provideSaveMetricsTask(
	useCase *usecase.SaveMetricsUseCase,
) (work.Task, error) {
	task, err := work.NewTask(
		"dump metrics",
		func(ctx context.Context) error {
			return useCase.Execute(ctx)
		},
	)
	task.SkipIfFullChan = true
	return task, err
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
