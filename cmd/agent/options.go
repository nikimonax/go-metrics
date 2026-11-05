package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"

	"github.com/nikimonax/go-metrics/internal/agent/config"
	"github.com/nikimonax/go-metrics/internal/lib/flagextra"
)

type Options struct {
	Address            string `env:"ADDRESS" json:"address"`
	APIVersion         uint   `env:"API" json:"api_version"`
	RateLimit          uint64 `env:"RATE_LIMIT" json:"rate_limit"`
	PollIntervalSecs   uint64 `env:"POLL_INTERVAL" json:"poll_interval"`
	ReportIntervalSecs uint64 `env:"REPORT_INTERVAL" json:"report_interval"`
	HashingKey         string `env:"KEY" json:"hashing_key"`
	CryptoKey          string `env:"CRYPTO_KEY" json:"crypto_key"`

	ConfigPath string `env:"CONFIG" json:"-"`
}

func (opts *Options) ToAgentConfig() (*config.AgentConfig, error) {
	cfg := config.NewDefaultConfig()

	if opts.Address != "" {
		rawURL := opts.Address

		hasScheme := false
		hasScheme = hasScheme || strings.HasPrefix(rawURL, "http://")
		hasScheme = hasScheme || strings.HasPrefix(rawURL, "https://")

		if !hasScheme {
			rawURL = "http://" + rawURL
		}

		baseURL, err := url.Parse(rawURL)

		if err != nil {
			return nil, fmt.Errorf("failed parse url '%s': %w", rawURL, err)
		}

		cfg.BaseURL = baseURL
	}

	if opts.APIVersion > 0 {
		cfg.APIVersion = opts.APIVersion
	}

	if opts.RateLimit > 0 {
		cfg.RateLimit = int64(opts.RateLimit)
	}

	if opts.PollIntervalSecs > 0 {
		cfg.PollInterval = time.Duration(opts.PollIntervalSecs) * time.Second
	}

	if opts.ReportIntervalSecs > 0 {
		cfg.ReportInterval = time.Duration(opts.ReportIntervalSecs) * time.Second
	}

	if opts.HashingKey != "" {
		cfg.Security.HashingKey = opts.HashingKey
	}

	if opts.CryptoKey != "" {
		cfg.Security.CryptoKeyFile = opts.CryptoKey
	}

	return &cfg, nil
}

func (opts *Options) Merge(other *Options) {
	if other.Address != "" {
		opts.Address = other.Address
	}

	if other.APIVersion > 0 {
		opts.APIVersion = other.APIVersion
	}

	if other.RateLimit > 0 {
		opts.RateLimit = other.RateLimit
	}

	if other.PollIntervalSecs > 0 {
		opts.PollIntervalSecs = other.PollIntervalSecs
	}

	if other.ReportIntervalSecs > 0 {
		opts.ReportIntervalSecs = other.ReportIntervalSecs
	}

	if other.HashingKey != "" {
		opts.HashingKey = other.HashingKey
	}

	if other.CryptoKey != "" {
		opts.CryptoKey = other.CryptoKey
	}

	if other.ConfigPath != "" {
		opts.ConfigPath = other.ConfigPath
	}
}

func ReadEnvOptions() (*Options, error) {
	var opts Options

	if err := env.Parse(&opts); err != nil {
		return nil, err
	}

	return &opts, nil
}

func ReadCliOptions() (*Options, error) {
	var opts Options

	cmd := flagextra.NewFlagSet()
	cmd.StringVar(
		&opts.Address,
		"a",
		"",
		"metrics server address",
	)
	cmd.UintVar(
		&opts.APIVersion,
		"v",
		0,
		"metrics server api version",
	)
	cmd.Uint64Var(
		&opts.RateLimit,
		"l",
		0,
		"requests rate limit",
	)
	cmd.Uint64Var(
		&opts.PollIntervalSecs,
		"p",
		0,
		"collect metrics interval",
	)
	cmd.Uint64Var(
		&opts.ReportIntervalSecs,
		"r",
		0,
		"send metrics interval",
	)
	cmd.StringVar(
		&opts.HashingKey,
		"k",
		"",
		"secret key to sign content",
	)
	cmd.StringVar(
		&opts.CryptoKey,
		"crypto-key",
		"",
		"public key file path to encrypt requests",
	)
	cmd.StringVar(
		&opts.ConfigPath,
		"config",
		"",
		"config file path",
	)
	cmd.StringVar(
		&opts.ConfigPath,
		"c",
		"",
		"config file path",
	)

	if err := cmd.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	return &opts, nil
}

func ReadFileOptions(filePath string) (*Options, error) {
	var opts Options

	content, err := os.ReadFile(filePath)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(content, &opts); err != nil {
		return nil, err
	}

	return &opts, nil
}

func ReadOptions() (*Options, error) {
	optionsFromEnv, err := ReadEnvOptions()

	if err != nil {
		return nil, err
	}

	optionsFromCli, err := ReadCliOptions()

	if err != nil {
		return nil, err
	}

	// priority: env -> cli -> file
	var opts = optionsFromCli
	opts.Merge(optionsFromEnv)

	if opts.ConfigPath != "" {
		optionsFromFile, err := ReadFileOptions(opts.ConfigPath)

		if err != nil {
			return nil, err
		}

		optionsFromFile.Merge(opts)
		opts = optionsFromFile
	}

	return opts, nil
}
