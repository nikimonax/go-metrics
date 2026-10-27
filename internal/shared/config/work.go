package config

import (
	"time"

	"go.uber.org/zap"

	"github.com/nikimonax/go-metrics/internal/lib/work"
)

const (
	DefaultStopTimeout = 5 * time.Second
)

func NewDefaultLifecycleConfig() work.LifecycleConfig {
	return work.LifecycleConfig{StopTimeout: DefaultStopTimeout}
}

func NewDefaultWorkerConfig(sugar *zap.SugaredLogger) work.WorkerConfig {
	return work.WorkerConfig{
		OnError: func(name string, err error) {
			sugar.Errorw("task failed", "task", name, "err", err)
		},
		OnPanic: func(name string, v any) {
			sugar.Errorw("task panic", "task", name, "value", v)
		},
	}
}

func NewDefaultSchedulerConfig(sugar *zap.SugaredLogger) work.SchedulerConfig {
	return work.SchedulerConfig{
		LifecycleConfig: NewDefaultLifecycleConfig(),
		OnError: func(name string, err error) {
			sugar.Errorw("failed submit task", "task", name, "err", err)
		},
	}
}
