package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env/v6"

	"github.com/nikimonax/go-metrics/internal/server/config"
)

type Options struct {
	Listen       string `env:"ADDRESS"`
	DumpFile     string `env:"FILE_STORAGE_PATH"`
	DumpInterval int64  `env:"STORE_INTERVAL" envDefault:"-1"`
	DumpRestore  *bool  `env:"RESTORE"`
}

func (opts *Options) ToServerConfig() *config.ServerConfig {
	cfg := config.NewDefaultConfig()

	if opts.Listen != "" {
		cfg.Listen = opts.Listen
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

	return &cfg
}

func (opts *Options) Merge(other *Options) {
	if other.Listen != "" {
		opts.Listen = other.Listen
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
}

func ReadEnvOptions() (*Options, error) {
	var opts Options

	if err := env.Parse(&opts); err != nil {
		return nil, err
	}

	return &opts, nil
}

func ReadCliOptions() (*Options, error) {
	var program string

	if len(os.Args) > 0 {
		program = os.Args[0]
	}

	cmd := flag.NewFlagSet(program, flag.ExitOnError)
	cmd.Usage = func() {
		_, _ = fmt.Fprintf(cmd.Output(), "Usage of %s:\n", program)
		cmd.PrintDefaults()
	}

	var (
		opts Options

		dumpRestore bool
	)

	cmd.StringVar(
		&opts.Listen,
		"a",
		"",
		"host and port to listen",
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

	// приоритет: env -> cli -> default
	optionsFromCli.Merge(optionsFromEnv)

	return optionsFromCli, nil
}
