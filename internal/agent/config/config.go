package config

import (
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"net/url"
	"time"

	"go.uber.org/zap"

	"github.com/nikimonax/go-metrics/internal/impl/gateway"
	"github.com/nikimonax/go-metrics/internal/lib/work"
	"github.com/nikimonax/go-metrics/internal/shared/config"
)

const (
	DefaultBaseURL        = "http://localhost:8080"
	DefaultAPIVersion     = 1
	DefaultPollInterval   = 2 * time.Second
	DefaultReportInterval = 10 * time.Second

	CompressLevel = gzip.BestSpeed
)

// some shared config re-exports
var (
	NewWorkerConfig    = config.NewDefaultWorkerConfig
	NewLifecycleConfig = config.NewDefaultLifecycleConfig
	NewSchedulerConfig = config.NewDefaultSchedulerConfig
)

type AgentConfig struct {
	BaseURL        *url.URL
	APIVersion     uint
	RateLimit      int64
	PollInterval   time.Duration
	ReportInterval time.Duration
	Backoff        config.BackoffConfig
	Security       config.SecurityConfig
}

func (cfg AgentConfig) Validate() error {
	if cfg.BaseURL == nil {
		return config.NewErrInvalidConfig("required non-nil 'BaseURL'")
	}

	maxAPIVersion := gateway.GetMaxAPIVersion()
	if cfg.APIVersion == 0 || cfg.APIVersion > maxAPIVersion {
		return config.NewErrInvalidConfig("required 'APIVersion' equals 1-" + fmt.Sprint(maxAPIVersion))
	}

	if cfg.RateLimit < 0 {
		return config.NewErrInvalidConfig("required 'RateLimit' greater or equals 0")
	}

	if cfg.PollInterval <= 0 {
		return config.NewErrInvalidConfig("required 'PollInterval' greater than 0")
	}

	if cfg.ReportInterval <= 0 {
		return config.NewErrInvalidConfig("required 'ReportInterval' greater than 0")
	}

	return nil
}

func NewDefaultConfig() AgentConfig {
	baseURL, err := url.Parse(DefaultBaseURL)

	if err != nil {
		err = fmt.Errorf(
			"failed parse default base url '%s': %w",
			DefaultBaseURL, err,
		)
		panic(err)
	}

	return AgentConfig{
		BaseURL:        baseURL,
		APIVersion:     DefaultAPIVersion,
		PollInterval:   DefaultPollInterval,
		ReportInterval: DefaultReportInterval,
		Backoff:        config.NewDefaultBackoffConfig(),
		Security: config.SecurityConfig{
			Header:      config.DefaultHashHeaderKey,
			HashingFunc: sha256.New,
		},
	}
}

func NewPoolConfig(sugar *zap.SugaredLogger) work.PoolConfig {
	return work.PoolConfig{
		WorkerConfig:    NewWorkerConfig(sugar),
		LifecycleConfig: NewLifecycleConfig(),
		// 2 workers and 2 queue size for collect and send tasks
		WorkerCount: 2,
		QueueSize:   2,
	}
}
