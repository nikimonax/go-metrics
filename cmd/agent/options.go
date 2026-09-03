package main

import (
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
	BaseURL            string `env:"ADDRESS"`
	APIVersion         uint   `env:"API"`
	PollIntervalSecs   uint64 `env:"POLL_INTERVAL"`
	ReportIntervalSecs uint64 `env:"REPORT_INTERVAL"`
}

func (opts *Options) ToAgentConfig() (*config.AgentConfig, error) {
	cfg := config.NewDefaultConfig()

	if opts.BaseURL != "" {
		rawURL := opts.BaseURL

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

	if opts.PollIntervalSecs > 0 {
		cfg.PollInterval = time.Duration(opts.PollIntervalSecs) * time.Second
	}

	if opts.ReportIntervalSecs > 0 {
		cfg.ReportInterval = time.Duration(opts.ReportIntervalSecs) * time.Second
	}

	return &cfg, nil
}

func (opts *Options) Merge(other *Options) {
	if other.BaseURL != "" {
		opts.BaseURL = other.BaseURL
	}

	if other.APIVersion > 0 {
		opts.APIVersion = other.APIVersion
	}

	if other.PollIntervalSecs > 0 {
		opts.PollIntervalSecs = other.PollIntervalSecs
	}

	if other.ReportIntervalSecs > 0 {
		opts.ReportIntervalSecs = other.ReportIntervalSecs
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
		&opts.BaseURL,
		"a",
		"",
		"metrics server base url",
	)
	cmd.UintVar(
		&opts.APIVersion,
		"v",
		0,
		"metrics server api version",
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

	if err := cmd.Parse(os.Args[1:]); err != nil {
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

	// приоритет: env -> cli -> default
	optionsFromCli.Merge(optionsFromEnv)

	return optionsFromCli, nil
}
