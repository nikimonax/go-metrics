package main

import (
	"flag"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/nikimonax/go-metrics/internal/agent"
)

const (
	defaultBaseURL            = "http://localhost:8080"
	defaultApiVersion         = 1
	defaultPollIntervalSecs   = 2
	defaultReportIntervalSecs = 10
)

type Options struct {
	BaseURL            string `env:"ADDRESS"`
	ApiVersion         uint   `env:"API"`
	PollIntervalSecs   uint64 `env:"POLL_INTERVAL"`
	ReportIntervalSecs uint64 `env:"REPORT_INTERVAL"`
}

func (opts *Options) ToAgentConfig() *agent.AgentConfig {
	rawUrl := opts.BaseURL

	hasScheme := false
	hasScheme = hasScheme || strings.HasPrefix(rawUrl, "http://")
	hasScheme = hasScheme || strings.HasPrefix(rawUrl, "https://")

	if !hasScheme {
		rawUrl = "http://" + rawUrl
	}

	baseUrl, err := url.Parse(rawUrl)

	if err != nil {
		log.Fatalf("failed parse url '%s': %s", rawUrl, err)
	}

	return &agent.AgentConfig{
		BaseURL:        baseUrl,
		ApiVersion:     opts.ApiVersion,
		PollInterval:   time.Duration(opts.PollIntervalSecs) * time.Second,
		ReportInterval: time.Duration(opts.ReportIntervalSecs) * time.Second,
	}
}

func (opts *Options) Merge(other Options) {
	if other.BaseURL != "" {
		opts.BaseURL = other.BaseURL
	}

	if other.ApiVersion > 0 {
		opts.ApiVersion = other.ApiVersion
	}

	if other.PollIntervalSecs > 0 {
		opts.PollIntervalSecs = other.PollIntervalSecs
	}

	if other.ReportIntervalSecs > 0 {
		opts.ReportIntervalSecs = other.ReportIntervalSecs
	}
}

func ReadOptions() *Options {
	var optionsFromEnv, optionsFromCli Options

	if err := env.Parse(&optionsFromEnv); err != nil {
		log.Fatalf("failed read env vars: %s", err)
	}

	flag.StringVar(
		&optionsFromCli.BaseURL,
		"a",
		defaultBaseURL,
		"metrics server base url",
	)
	flag.UintVar(
		&optionsFromCli.ApiVersion,
		"v",
		defaultApiVersion,
		"metrics server api version",
	)
	flag.Uint64Var(
		&optionsFromCli.PollIntervalSecs,
		"p",
		defaultPollIntervalSecs,
		"collect metrics interval",
	)
	flag.Uint64Var(
		&optionsFromCli.ReportIntervalSecs,
		"r",
		defaultReportIntervalSecs,
		"send metrics interval",
	)
	flag.Parse()

	// приоритет: env -> cli -> default
	optionsFromCli.Merge(optionsFromEnv)

	return &optionsFromCli
}
