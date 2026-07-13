package main

import (
	"flag"
	"log"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/nikimonax/go-metrics/internal/agent"
)

const (
	defaultBaseURL            = "http://localhost:8080"
	defaultPollIntervalSecs   = 2
	defaultReportIntervalSecs = 10
)

type Options struct {
	BaseURL            string `env:"ADDRESS"`
	PollIntervalSecs   uint64 `env:"POLL_INTERVAL"`
	ReportIntervalSecs uint64 `env:"REPORT_INTERVAL"`
}

func (opts *Options) ToAgentConfig() *agent.AgentConfig {
	baseUrl := opts.BaseURL

	hasScheme := false
	hasScheme = hasScheme || strings.HasPrefix(baseUrl, "http://")
	hasScheme = hasScheme || strings.HasPrefix(baseUrl, "https://")

	if !hasScheme {
		baseUrl = "http://" + baseUrl
	}

	return &agent.AgentConfig{
		BaseURL:        baseUrl,
		PollInterval:   time.Duration(opts.PollIntervalSecs) * time.Second,
		ReportInterval: time.Duration(opts.ReportIntervalSecs) * time.Second,
	}
}

func (opts *Options) Merge(other Options) {
	if other.BaseURL != "" {
		opts.BaseURL = other.BaseURL
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
