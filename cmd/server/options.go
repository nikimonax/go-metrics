package main

import (
	"encoding/json"
	"flag"
	"os"
	"time"

	"github.com/caarlos0/env/v6"

	"github.com/nikimonax/go-metrics/internal/lib/flagextra"
	"github.com/nikimonax/go-metrics/internal/server/config"
)

type Options struct {
	Address      string  `env:"ADDRESS"           json:"address"`
	Profile      *bool   `env:"PROFILE"           json:"profile"`
	DatabaseDSN  string  `env:"DATABASE_DSN"      json:"database_dsn"`
	HashingKey   string  `env:"KEY"               json:"hashing_key"`
	CryptoKey    string  `env:"CRYPTO_KEY"        json:"crypto_key"`
	DumpFile     string  `env:"FILE_STORAGE_PATH" json:"store_file"`
	DumpInterval *uint64 `env:"STORE_INTERVAL"    json:"store_interval"`
	DumpRestore  *bool   `env:"RESTORE"           json:"restore"`
	AuditFile    string  `env:"AUDIT_FILE"        json:"audit_file"`
	AuditURL     string  `env:"AUDIT_URL"         json:"audit_url"`

	ConfigPath string `env:"CONFIG" json:"-"`
}

func (opts *Options) ToServerConfig() *config.ServerConfig {
	cfg := config.NewDefaultConfig()

	if opts.Address != "" {
		cfg.Listen = opts.Address
	}

	if opts.Profile != nil && *opts.Profile {
		cfg.Profile = true
	}

	if opts.DatabaseDSN != "" {
		cfg.Database.DSN = opts.DatabaseDSN
	}

	if opts.HashingKey != "" {
		cfg.Security.HashingKey = opts.HashingKey
	}

	if opts.CryptoKey != "" {
		cfg.Security.CryptoKeyFile = opts.CryptoKey
	}

	if opts.DumpFile != "" {
		cfg.Dump.File = opts.DumpFile
	}

	if opts.DumpInterval != nil {
		cfg.Dump.Interval = time.Duration(*opts.DumpInterval) * time.Second
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
	if other.Address != "" {
		opts.Address = other.Address
	}

	if other.Profile != nil {
		opts.Profile = other.Profile
	}

	if other.DatabaseDSN != "" {
		opts.DatabaseDSN = other.DatabaseDSN
	}

	if other.HashingKey != "" {
		opts.HashingKey = other.HashingKey
	}

	if other.CryptoKey != "" {
		opts.CryptoKey = other.CryptoKey
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

	if other.AuditFile != "" {
		opts.AuditFile = other.AuditFile
	}

	if other.AuditURL != "" {
		opts.AuditURL = other.AuditURL
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
	var (
		opts Options

		dumpInterval uint64
		dumpRestore  bool
		profile      bool
	)

	cmd := flagextra.NewFlagSet()
	cmd.StringVar(
		&opts.Address,
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
		&opts.CryptoKey,
		"crypto-key",
		"",
		"private key file path to decrypt requests",
	)
	cmd.StringVar(
		&opts.DumpFile,
		"f",
		"",
		"file path to dump metrics",
	)
	cmd.Uint64Var(
		&dumpInterval,
		"i",
		0,
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
	cmd.BoolVar(
		&profile,
		"pprof",
		false,
		"enable profiling",
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

	cmd.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "i":
			opts.DumpInterval = &dumpInterval
		case "r":
			opts.DumpRestore = &dumpRestore
		case "pprof":
			opts.Profile = &profile
		}
	})

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
