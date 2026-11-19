package config

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net/netip"
	"net/url"
	"time"

	"go.uber.org/zap/zapcore"

	"github.com/nikimonax/go-metrics/internal/lib/zapextra"
	"github.com/nikimonax/go-metrics/internal/shared/config"
)

const (
	DefaultListen          = "localhost:8080"
	DefaultDatabaseMigrate = true
	DefaultDumpFile        = "metrics.json"
	DefaultDumpInterval    = 300 * time.Second
	DefaultDumpRestore     = false

	SourceIPHeader = config.SourceIPKey
)

// some shared config re-exports
var (
	ErrInvalidConfig    = config.ErrInvalidConfig
	NewErrInvalidConfig = config.NewErrInvalidConfig

	NewWorkerConfig    = config.NewDefaultWorkerConfig
	NewLifecycleConfig = config.NewDefaultLifecycleConfig
	NewSchedulerConfig = config.NewDefaultSchedulerConfig
)

type ServerType string

const (
	GRPCServer ServerType = "grpc"
	HTTPServer ServerType = "http"
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
		return NewErrInvalidConfig(msg)
	}

	if dsnURL.Scheme != "" && dsnURL.Scheme != "postgres" {
		msg := fmt.Sprintf("unsupported database: %s", dsnURL.Scheme)
		return NewErrInvalidConfig(msg)
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
		err := NewErrInvalidConfig("required 'File'")
		errs = errors.Join(err, errs)
	}

	if cfg.Interval < 0 {
		err := NewErrInvalidConfig("required non negative 'Interval'")
		errs = errors.Join(err, errs)
	}

	return errs
}

type AuditConfig struct {
	File    string
	URL     string
	Backoff config.BackoffConfig
}

func (cfg AuditConfig) Validate() error {
	if cfg.URL == "" {
		return nil
	}

	url, err := url.Parse(cfg.URL)

	if err != nil {
		msg := fmt.Sprintf("invalid audit url: %s", err)
		return NewErrInvalidConfig(msg)
	}

	if url.Scheme != "http" && url.Scheme != "https" {
		msg := fmt.Sprintf("unsupported audit url scheme: %s", url.Scheme)
		return NewErrInvalidConfig(msg)
	}

	return nil
}

type ServerConfig struct {
	Listen        string
	ServerType    ServerType
	TrustedSubnet string
	Profile       bool
	Database      DatabaseConfig
	Security      config.SecurityConfig
	Log           LogConfig
	Dump          DumpConfig
	Audit         AuditConfig
}

func (cfg ServerConfig) Validate() error {
	var errs error

	if cfg.Listen == "" {
		errs = NewErrInvalidConfig("required 'Listen'")
	}

	if cfg.ServerType != HTTPServer && cfg.ServerType != GRPCServer {
		msg := fmt.Sprintf(
			"invalid 'ServerType', expected 'http' or 'grpc', but got '%s'",
			cfg.ServerType,
		)
		err := NewErrInvalidConfig(msg)
		errs = errors.Join(errs, err)
	}

	if cfg.TrustedSubnet != "" {
		_, err := netip.ParsePrefix(cfg.TrustedSubnet)

		if err != nil {
			err = NewErrInvalidConfig(
				fmt.Sprintf("invalid 'TrustedSubnet': %s", err),
			)
		}

		errs = errors.Join(errs, err)
	}

	return errors.Join(
		errs,
		cfg.Database.Validate(),
		cfg.Dump.Validate(),
		cfg.Audit.Validate(),
	)
}

func NewDefaultConfig() ServerConfig {
	return ServerConfig{
		Listen:     DefaultListen,
		ServerType: HTTPServer,
		Database: DatabaseConfig{
			Migrate: DefaultDatabaseMigrate,
			Backoff: config.NewDefaultBackoffConfig(),
		},
		Security: config.SecurityConfig{
			Header:      config.DefaultHashHeaderKey,
			HashingFunc: sha256.New,
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
		Audit: AuditConfig{
			Backoff: config.NewDefaultBackoffConfig(),
		},
	}
}
