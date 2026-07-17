package main

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/nikimonax/go-metrics/internal/server"
)

const defaultBaseURL = "localhost:8080"

type Options struct {
	BaseURL string `env:"ADDRESS"`
}

func (opts *Options) ToServerConfig() *server.ServerConfig {
	return &server.ServerConfig{BaseURL: opts.BaseURL}
}

func (opts *Options) Merge(other Options) {
	if other.BaseURL != "" {
		opts.BaseURL = other.BaseURL
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
		"host and port to listen",
	)
	flag.Parse()

	// приоритет: env -> cli -> default
	optionsFromCli.Merge(optionsFromEnv)

	return &optionsFromCli
}
