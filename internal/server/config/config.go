package config

import (
	"time"

	"go.uber.org/zap/zapcore"

	"github.com/nikimonax/go-metrics/internal/lib/work"
	"github.com/nikimonax/go-metrics/internal/lib/zapextra"
	"github.com/nikimonax/go-metrics/internal/shared/config"
)

const (
	DefaultListen       = "localhost:8080"
	DefaultDumpFile     = "metrics.json"
	DefaultDumpInterval = 300 * time.Second
	DefaultDumpRestore  = false
	DefaultStopTimeout  = 5 * time.Second
)

type LogConfig struct {
	Env   zapextra.LogEnv
	Level zapcore.Level
}

type DumpConfig struct {
	File     string
	Interval time.Duration
	Restore  bool
}

func (cfg DumpConfig) Validate() error {
	if cfg.File == "" {
		return config.NewErrInvalidConfig("required 'File'")
	}

	if cfg.Interval < 0 {
		return config.NewErrInvalidConfig("required non negative 'Interval'")
	}

	return nil
}

type ServerConfig struct {
	Listen    string
	Log       LogConfig
	Dump      DumpConfig
	Pool      work.PoolConfig
	Scheduler work.SchedulerConfig
}

func (cfg ServerConfig) Validate() error {
	if cfg.Listen == "" {
		return config.NewErrInvalidConfig("required 'Listen'")
	}

	if err := cfg.Dump.Validate(); err != nil {
		return err
	}

	if err := cfg.Pool.Validate(); err != nil {
		return err
	}

	if err := cfg.Scheduler.Validate(); err != nil {
		return err
	}

	return nil
}

func NewDefaultConfig() ServerConfig {
	return ServerConfig{
		Listen: DefaultListen,
		Log: LogConfig{
			Env:   zapextra.EnvDev,
			Level: zapcore.InfoLevel,
		},
		Dump: DumpConfig{
			File:     DefaultDumpFile,
			Interval: DefaultDumpInterval,
			Restore:  DefaultDumpRestore,
		},
		Pool: work.PoolConfig{
			LifecycleConfig: work.LifecycleConfig{
				StopTimeout: DefaultStopTimeout,
			},
			WorkerCount: 1,
			QueueSize:   20,
		},
		Scheduler: work.SchedulerConfig{
			LifecycleConfig: work.LifecycleConfig{
				StopTimeout: DefaultStopTimeout,
			},
		},
	}
}
