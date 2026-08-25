package fxmodule

import (
	"context"
	"net/http"

	"go.uber.org/fx"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/app/usecase"
	"github.com/nikimonax/go-metrics/internal/impl/dumper"
	"github.com/nikimonax/go-metrics/internal/impl/serializer"
	"github.com/nikimonax/go-metrics/internal/lib/work"
	"github.com/nikimonax/go-metrics/internal/server/config"
	mymiddleware "github.com/nikimonax/go-metrics/internal/server/middleware"
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
		fx.Invoke(
			registerRestore,
			registerSyncDumps,
			registerPeriodicDumps,
		),
	)
}

func provideMetricDumper(
	cfg *config.ServerConfig,
	serializer serializer.MetricSerializer,
) interfaces.MetricDumper {
	return dumper.NewFileMetricDumper(cfg.Dump.File, serializer)
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
	updateMetricsHook *mymiddleware.RequestHook,
	saveMetricsUseCase *usecase.SaveMetricsUseCase,
) {
	if cfg.Dump.Interval != 0 {
		return
	}

	updateMetricsHook.AfterRequest(
		func(_ *http.Request) error {
			return saveMetricsUseCase.Execute()
		},
	)
}

func registerPeriodicDumps(
	cfg *config.ServerConfig,
	scheduler *work.Scheduler,
	saveMetricsUseCase *usecase.SaveMetricsUseCase,
) error {
	if cfg.Dump.Interval <= 0 {
		return nil
	}

	task, err := work.NewTask(
		"dump metrics",
		func(_ context.Context) error {
			return saveMetricsUseCase.Execute()
		},
	)

	if err != nil {
		return err
	}

	return scheduler.Schedule(task, cfg.Dump.Interval)
}
