package config

import (
	"compress/gzip"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net"
	"net/url"
	"time"

	"go.uber.org/zap"

	"github.com/nikimonax/go-metrics/internal/impl/gateway"
	"github.com/nikimonax/go-metrics/internal/lib/netutil"
	"github.com/nikimonax/go-metrics/internal/lib/work"
	"github.com/nikimonax/go-metrics/internal/shared/config"
)

const (
	DefaultBaseURL        = "http://localhost:8080"
	DefaultAPIVersion     = 1
	DefaultPollInterval   = 2 * time.Second
	DefaultReportInterval = 10 * time.Second

	SourceIPHeader = config.SourceIPKey
	CompressLevel  = gzip.BestSpeed
)

// some shared config re-exports
var (
	ErrInvalidConfig    = config.ErrInvalidConfig
	NewErrInvalidConfig = config.NewErrInvalidConfig

	NewWorkerConfig    = config.NewDefaultWorkerConfig
	NewLifecycleConfig = config.NewDefaultLifecycleConfig
	NewSchedulerConfig = config.NewDefaultSchedulerConfig
)

func TryGetDefaultSourceIP(host string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)

	if err != nil {
		return "", err
	}

	sourceIP, err := netutil.GetSourceIPFor(addrs[0].IP)

	if err != nil {
		if errors.Is(err, netutil.ErrExternalTarget) {
			return "", nil
		}

		return "", err
	}

	return sourceIP.String(), nil
}

type AgentConfig struct {
	BaseURL        *url.URL
	GrpcMode       bool
	SourceIP       string
	APIVersion     uint
	RateLimit      int64
	PollInterval   time.Duration
	ReportInterval time.Duration
	Backoff        config.BackoffConfig
	Security       config.SecurityConfig
}

func (cfg AgentConfig) Validate() error {
	var errs error

	if cfg.BaseURL == nil {
		err := NewErrInvalidConfig("required non-nil 'BaseURL'")
		errs = errors.Join(errs, err)
	}

	if cfg.SourceIP != "" && net.ParseIP(cfg.SourceIP) == nil {
		err := NewErrInvalidConfig("invalid 'SourceIP' address")
		errs = errors.Join(errs, err)
	}

	maxAPIVersion := gateway.GetMaxAPIVersion()
	if cfg.APIVersion == 0 || cfg.APIVersion > maxAPIVersion {
		err := NewErrInvalidConfig(
			"required 'APIVersion' equals 1-" + fmt.Sprint(maxAPIVersion),
		)
		errs = errors.Join(errs, err)
	}

	if cfg.RateLimit < 0 {
		err := NewErrInvalidConfig("required 'RateLimit' greater or equals 0")
		errs = errors.Join(errs, err)
	}

	if cfg.PollInterval <= 0 {
		err := NewErrInvalidConfig("required 'PollInterval' greater than 0")
		errs = errors.Join(errs, err)
	}

	if cfg.ReportInterval <= 0 {
		err := NewErrInvalidConfig("required 'ReportInterval' greater than 0")
		errs = errors.Join(errs, err)
	}

	return errs
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
