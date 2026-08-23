package main

import (
	"flag"
	"log"
	"time"

	"github.com/caarlos0/env/v6"

	"github.com/nikimonax/go-metrics/internal/server/config"
)

const (
	defaultBaseURL      = "localhost:8080"
	defaultDumpFile     = "metrics.json"
	defaultDumpInterval = 300
	defaultDumpRestore  = false
)

type Options struct {
	BaseURL      string  `env:"ADDRESS"`
	DumpFile     string  `env:"FILE_STORAGE_PATH"`
	DumpInterval *uint64 `env:"STORE_INTERVAL"`
	DumpRestore  *bool   `env:"RESTORE"`
}

func (opts *Options) ToServerConfig() *config.ServerConfig {
	var (
		DumpInterval = defaultDumpInterval * time.Second
		DumpRestore  = defaultDumpRestore
	)

	if opts.DumpInterval != nil {
		DumpInterval = time.Duration(*opts.DumpInterval) * time.Second
	}

	if opts.DumpRestore != nil {
		DumpRestore = *opts.DumpRestore
	}

	return &config.ServerConfig{
		BaseURL:              opts.BaseURL,
		DumpFile:             opts.DumpFile,
		DumpInterval:         DumpInterval,
		DumpRestore:          DumpRestore,
		LifespanCloseTimeout: config.DefaultLifespanCloseTimeout,
		ServerStopTimeout:    config.DefaultServerStopTimeout,
	}
}

func (opts *Options) Merge(other Options) {
	if other.BaseURL != "" {
		opts.BaseURL = other.BaseURL
	}

	if other.DumpFile != "" {
		opts.DumpFile = other.DumpFile
	}

	if other.DumpInterval != nil {
		opts.DumpInterval = other.DumpInterval
	}

	if other.DumpRestore != nil {
		opts.DumpRestore = other.DumpRestore
	}
}

func ReadOptions() *Options {
	var optionsFromEnv, optionsFromCli Options

	if err := env.Parse(&optionsFromEnv); err != nil {
		log.Fatalf("failed read env vars: %s", err)
	}

	var (
		DumpInterval uint64
		DumpRestore  bool
	)

	optionsFromCli.DumpInterval = &DumpInterval
	optionsFromCli.DumpRestore = &DumpRestore

	flag.StringVar(
		&optionsFromCli.BaseURL,
		"a",
		defaultBaseURL,
		"host and port to listen",
	)
	flag.StringVar(
		&optionsFromCli.DumpFile,
		"f",
		defaultDumpFile,
		"file path to dump metrics",
	)
	flag.Uint64Var(
		&DumpInterval,
		"i",
		defaultDumpInterval,
		"time interval to dump metrics",
	)
	flag.BoolVar(
		&DumpRestore,
		"r",
		defaultDumpRestore,
		"restore metrics from the dump file at startup",
	)
	flag.Parse()

	// приоритет: env -> cli -> default
	optionsFromCli.Merge(optionsFromEnv)

	return &optionsFromCli
}
