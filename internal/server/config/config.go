package config

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"go.uber.org/zap/zapcore"

	"github.com/nikimonax/go-metrics/internal/lib/work"
	"github.com/nikimonax/go-metrics/internal/lib/zapextra"
	"github.com/nikimonax/go-metrics/internal/shared/config"
)

const (
	DefaultListen          = "localhost:8080"
	DefaultDatabaseDsn     = ""
	DefaultDatabaseMigrate = true
	DefaultDumpFile        = "metrics.json"
	DefaultDumpInterval    = 300 * time.Second
	DefaultDumpRestore     = false
	DefaultStopTimeout     = 5 * time.Second
)

type DatabaseConfig struct {
	DSN     string
	Migrate bool
	Backoff config.BackoffConfig
}

func (cfg DatabaseConfig) Validate() error {
	if cfg.DSN == "" {
		return nil
	}

	dsnURL, err := url.Parse(cfg.DSN)

	if err != nil {
		msg := fmt.Sprintf("invalid database dsn: %s", err)
		return config.NewErrInvalidConfig(msg)
	}

	if dsnURL.Scheme != "" && dsnURL.Scheme != "postgres" {
		msg := fmt.Sprintf("unsupported database: %s", dsnURL.Scheme)
		return config.NewErrInvalidConfig(msg)
	}

	return nil
}

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
	var errs error

	if cfg.File == "" {
		err := config.NewErrInvalidConfig("required 'File'")
		errs = errors.Join(err, errs)
	}

	if cfg.Interval < 0 {
		err := config.NewErrInvalidConfig("required non negative 'Interval'")
		errs = errors.Join(err, errs)
	}

	return errs
}

type ServerConfig struct {
	Listen    string
	Database  DatabaseConfig
	Log       LogConfig
	Dump      DumpConfig
	Pool      work.PoolConfig
	Scheduler work.SchedulerConfig
}

func (cfg ServerConfig) Validate() error {
	var errs error

	if cfg.Listen == "" {
		err := config.NewErrInvalidConfig("required 'Listen'")
		errs = errors.Join(err, errs)
	}

	if err := cfg.Database.Validate(); err != nil {
		errs = errors.Join(err, errs)
	}

	if err := cfg.Dump.Validate(); err != nil {
		errs = errors.Join(err, errs)
	}

	if err := cfg.Pool.Validate(); err != nil {
		errs = errors.Join(err, errs)
	}

	if err := cfg.Scheduler.Validate(); err != nil {
		errs = errors.Join(err, errs)
	}

	return errs
}

func NewDefaultConfig() ServerConfig {
	return ServerConfig{
		Listen: DefaultListen,
		Database: DatabaseConfig{
			DSN:     DefaultDatabaseDsn,
			Migrate: DefaultDatabaseMigrate,
			Backoff: config.BackoffConfig{
				Retry: 3,
				Seed:  1 * time.Second,
				Add:   2 * time.Second,
			},
		},
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
			QueueSize:   1,
		},
		Scheduler: work.SchedulerConfig{
			LifecycleConfig: work.LifecycleConfig{
				StopTimeout: DefaultStopTimeout,
			},
		},
	}
}
