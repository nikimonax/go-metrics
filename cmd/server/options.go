package main

import (
	"flag"
	"os"
	"time"

	"github.com/caarlos0/env/v6"

	"github.com/nikimonax/go-metrics/internal/lib/flagextra"
	"github.com/nikimonax/go-metrics/internal/server/config"
)

type Options struct {
	Listen       string `env:"ADDRESS"`
	DatabaseDSN  string `env:"DATABASE_DSN"`
	HashingKey   string `env:"KEY"`
	DumpFile     string `env:"FILE_STORAGE_PATH"`
	DumpInterval int64  `env:"STORE_INTERVAL" envDefault:"-1"`
	DumpRestore  *bool  `env:"RESTORE"`
	AuditFile    string `env:"AUDIT_FILE"`
	AuditURL     string `env:"AUDIT_URL"`
}

func (opts *Options) ToServerConfig() *config.ServerConfig {
	cfg := config.NewDefaultConfig()

	if opts.Listen != "" {
		cfg.Listen = opts.Listen
	}

	if opts.DatabaseDSN != "" {
		cfg.Database.DSN = opts.DatabaseDSN
	}

	if opts.HashingKey != "" {
		cfg.Security.HashingKey = opts.HashingKey
	}

	if opts.DumpFile != "" {
		cfg.Dump.File = opts.DumpFile
	}

	if opts.DumpInterval >= 0 {
		cfg.Dump.Interval = time.Duration(opts.DumpInterval) * time.Second
	}

	if opts.DumpRestore != nil && *opts.DumpRestore {
		cfg.Dump.Restore = true
	}

	if opts.AuditFile != "" {
		cfg.Audit.File = opts.AuditFile
	}

	if opts.AuditURL != "" {
		cfg.Audit.URL = opts.AuditURL
	}

	return &cfg
}

func (opts *Options) Merge(other *Options) {
	if other.Listen != "" {
		opts.Listen = other.Listen
	}

	if other.DatabaseDSN != "" {
		opts.DatabaseDSN = other.DatabaseDSN
	}

	if other.HashingKey != "" {
		opts.HashingKey = other.HashingKey
	}

	if other.DumpFile != "" {
		opts.DumpFile = other.DumpFile
	}

	if other.DumpInterval >= 0 {
		opts.DumpInterval = other.DumpInterval
	}

	if other.DumpRestore != nil {
		opts.DumpRestore = other.DumpRestore
	}

	if other.AuditFile != "" {
		opts.AuditFile = other.AuditFile
	}

	if other.AuditURL != "" {
		opts.AuditURL = other.AuditURL
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
	var (
		opts Options

		dumpRestore bool
	)

	cmd := flagextra.NewFlagSet()
	cmd.StringVar(
		&opts.Listen,
		"a",
		"",
		"host and port to listen",
	)
	cmd.StringVar(
		&opts.DatabaseDSN,
		"d",
		"",
		"database dsn for connection",
	)
	cmd.StringVar(
		&opts.HashingKey,
		"k",
		"",
		"secret key to verify requests and calculate response hash",
	)
	cmd.StringVar(
		&opts.DumpFile,
		"f",
		"",
		"file path to dump metrics",
	)
	cmd.Int64Var(
		&opts.DumpInterval,
		"i",
		-1,
		"time interval to dump metrics",
	)
	cmd.BoolVar(
		&dumpRestore,
		"r",
		false,
		"restore metrics from the dump file at startup",
	)
	cmd.StringVar(
		&opts.AuditFile,
		"audit-file",
		"",
		"file path to write audit events",
	)
	cmd.StringVar(
		&opts.AuditURL,
		"audit-url",
		"",
		"http endpoint to send metrics there",
	)

	if err := cmd.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	cmd.Visit(func(f *flag.Flag) {
		if f.Name == "r" {
			opts.DumpRestore = &dumpRestore
		}
	})

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

	// priority: env -> cli -> default
	optionsFromCli.Merge(optionsFromEnv)

	return optionsFromCli, nil
}
