package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/go-metrics/internal/server/config"
)

func withArgs(t *testing.T, args ...string) {
	t.Helper()

	original := os.Args
	os.Args = append([]string{original[0]}, args...)

	t.Cleanup(func() {
		os.Args = original
	})
}

// env v6 ignores empty values, so this neutralizes ambient variables
func clearEnv(t *testing.T, keys ...string) {
	t.Helper()

	for _, key := range keys {
		t.Setenv(key, "")
	}
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.json")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	return path
}

func boolPtr(v bool) *bool {
	return &v
}

func uint64Ptr(v uint64) *uint64 {
	return &v
}

func TestToServerConfig(t *testing.T) {
	tests := []struct {
		name  string
		opts  Options
		check func(t *testing.T, cfg *config.ServerConfig)
	}{
		{
			name: "empty options: defaults kept",
			check: func(t *testing.T, cfg *config.ServerConfig) {
				want := config.NewDefaultConfig()

				assert.Equal(t, want.Listen, cfg.Listen)
				assert.False(t, cfg.Profile)
				assert.Empty(t, cfg.Database.DSN)
				assert.Empty(t, cfg.Security.HashingKey)
				assert.Empty(t, cfg.Security.CryptoKeyFile)
				assert.Equal(t, want.Dump, cfg.Dump)
				assert.Empty(t, cfg.Audit.File)
				assert.Empty(t, cfg.Audit.URL)
			},
		},
		{
			name: "all fields set",
			opts: Options{
				Address:      "localhost:9090",
				Profile:      boolPtr(true),
				DatabaseDSN:  "postgres://localhost:5432/metrics",
				HashingKey:   "secret",
				CryptoKey:    "private_key.pem",
				DumpFile:     "dump.json",
				DumpInterval: uint64Ptr(5),
				DumpRestore:  boolPtr(true),
				AuditFile:    "audit.json",
				AuditURL:     "http://localhost:9091/audit",
			},
			check: func(t *testing.T, cfg *config.ServerConfig) {
				assert.Equal(t, "localhost:9090", cfg.Listen)
				assert.True(t, cfg.Profile)
				assert.Equal(t, "postgres://localhost:5432/metrics", cfg.Database.DSN)
				assert.Equal(t, "secret", cfg.Security.HashingKey)
				assert.Equal(t, "private_key.pem", cfg.Security.CryptoKeyFile)
				assert.Equal(t, "dump.json", cfg.Dump.File)
				assert.Equal(t, 5*time.Second, cfg.Dump.Interval)
				assert.True(t, cfg.Dump.Restore)
				assert.Equal(t, "audit.json", cfg.Audit.File)
				assert.Equal(t, "http://localhost:9091/audit", cfg.Audit.URL)
			},
		},
		{
			name: "profile false pointer does not enable profiling",
			opts: Options{Profile: boolPtr(false)},
			check: func(t *testing.T, cfg *config.ServerConfig) {
				assert.False(t, cfg.Profile)
			},
		},
		{
			name: "restore false pointer does not enable restore",
			opts: Options{DumpRestore: boolPtr(false)},
			check: func(t *testing.T, cfg *config.ServerConfig) {
				assert.False(t, cfg.Dump.Restore)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := tc.opts.ToServerConfig()

			require.NotNil(t, cfg)
			tc.check(t, cfg)
		})
	}
}

func TestServerOptionsMerge(t *testing.T) {
	tests := []struct {
		name  string
		base  Options
		other Options
		want  Options
	}{
		{
			name: "empty other keeps base",
			base: Options{
				Address:      "base:1111",
				Profile:      boolPtr(true),
				DatabaseDSN:  "dsn",
				HashingKey:   "key",
				CryptoKey:    "crypto",
				DumpFile:     "dump.json",
				DumpInterval: uint64Ptr(5),
				DumpRestore:  boolPtr(true),
				AuditFile:    "audit.json",
				AuditURL:     "http://audit",
				ConfigPath:   "cfg.json",
			},
			other: Options{},
			want: Options{
				Address:      "base:1111",
				Profile:      boolPtr(true),
				DatabaseDSN:  "dsn",
				HashingKey:   "key",
				CryptoKey:    "crypto",
				DumpFile:     "dump.json",
				DumpInterval: uint64Ptr(5),
				DumpRestore:  boolPtr(true),
				AuditFile:    "audit.json",
				AuditURL:     "http://audit",
				ConfigPath:   "cfg.json",
			},
		},
		{
			name: "non-empty other overrides",
			base: Options{Address: "base:1111", Profile: boolPtr(false)},
			other: Options{
				Address:      "other:2222",
				Profile:      boolPtr(true),
				DatabaseDSN:  "dsn",
				HashingKey:   "key",
				CryptoKey:    "crypto",
				DumpFile:     "dump.json",
				DumpInterval: uint64Ptr(5),
				DumpRestore:  boolPtr(true),
				AuditFile:    "audit.json",
				AuditURL:     "http://audit",
				ConfigPath:   "cfg.json",
			},
			want: Options{
				Address:      "other:2222",
				Profile:      boolPtr(true),
				DatabaseDSN:  "dsn",
				HashingKey:   "key",
				CryptoKey:    "crypto",
				DumpFile:     "dump.json",
				DumpInterval: uint64Ptr(5),
				DumpRestore:  boolPtr(true),
				AuditFile:    "audit.json",
				AuditURL:     "http://audit",
				ConfigPath:   "cfg.json",
			},
		},
		{
			name: "nil pointers do not override",
			base: Options{
				Profile:      boolPtr(true),
				DumpInterval: uint64Ptr(5),
				DumpRestore:  boolPtr(true),
			},
			other: Options{Address: "other:2222"},
			want: Options{
				Address:      "other:2222",
				Profile:      boolPtr(true),
				DumpInterval: uint64Ptr(5),
				DumpRestore:  boolPtr(true),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			opts := tc.base
			opts.Merge(&tc.other)

			assert.Equal(t, tc.want, opts)
		})
	}
}

func TestServerReadEnvOptions(t *testing.T) {
	t.Run("empty environment", func(t *testing.T) {
		clearEnv(t,
			"ADDRESS", "PROFILE", "DATABASE_DSN",
			"KEY", "CRYPTO_KEY", "FILE_STORAGE_PATH",
			"STORE_INTERVAL", "RESTORE", "AUDIT_FILE", "AUDIT_URL", "CONFIG",
		)

		opts, err := ReadEnvOptions()

		require.NoError(t, err)
		assert.Equal(t, &Options{}, opts)
	})

	t.Run("all variables set", func(t *testing.T) {
		t.Setenv("ADDRESS", "localhost:9090")
		t.Setenv("PROFILE", "true")
		t.Setenv("DATABASE_DSN", "postgres://localhost:5432/metrics")
		t.Setenv("KEY", "secret")
		t.Setenv("CRYPTO_KEY", "private_key.pem")
		t.Setenv("FILE_STORAGE_PATH", "dump.json")
		t.Setenv("STORE_INTERVAL", "5")
		t.Setenv("RESTORE", "true")
		t.Setenv("AUDIT_FILE", "audit.json")
		t.Setenv("AUDIT_URL", "http://localhost:9091/audit")

		opts, err := ReadEnvOptions()

		require.NoError(t, err)
		assert.Equal(t, &Options{
			Address:      "localhost:9090",
			Profile:      boolPtr(true),
			DatabaseDSN:  "postgres://localhost:5432/metrics",
			HashingKey:   "secret",
			CryptoKey:    "private_key.pem",
			DumpFile:     "dump.json",
			DumpInterval: uint64Ptr(5),
			DumpRestore:  boolPtr(true),
			AuditFile:    "audit.json",
			AuditURL:     "http://localhost:9091/audit",
		}, opts)
	})

	t.Run("invalid uint value", func(t *testing.T) {
		t.Setenv("STORE_INTERVAL", "not-a-number")

		_, err := ReadEnvOptions()

		assert.Error(t, err)
	})
}

func TestServerReadCliOptions(t *testing.T) {
	t.Run("no flags", func(t *testing.T) {
		withArgs(t)

		opts, err := ReadCliOptions()

		require.NoError(t, err)
		assert.Equal(t, &Options{}, opts)
	})

	t.Run("all flags", func(t *testing.T) {
		withArgs(t,
			"-a", "localhost:9090",
			"-d", "postgres://localhost:5432/metrics",
			"-k", "secret",
			"-crypto-key", "private_key.pem",
			"-f", "dump.json",
			"-i", "5",
			"-r",
			"-audit-file", "audit.json",
			"-audit-url", "http://localhost:9091/audit",
			"-pprof",
			"-config", "config.json",
		)

		opts, err := ReadCliOptions()

		require.NoError(t, err)
		assert.Equal(t, &Options{
			Address:      "localhost:9090",
			DatabaseDSN:  "postgres://localhost:5432/metrics",
			HashingKey:   "secret",
			CryptoKey:    "private_key.pem",
			DumpFile:     "dump.json",
			DumpInterval: uint64Ptr(5),
			DumpRestore:  boolPtr(true),
			AuditFile:    "audit.json",
			AuditURL:     "http://localhost:9091/audit",
			Profile:      boolPtr(true),
			ConfigPath:   "config.json",
		}, opts)
	})

	t.Run("pointer flags not passed stay nil", func(t *testing.T) {
		withArgs(t, "-a", "localhost:9090")

		opts, err := ReadCliOptions()

		require.NoError(t, err)
		assert.Nil(t, opts.Profile)
		assert.Nil(t, opts.DumpInterval)
		assert.Nil(t, opts.DumpRestore)
	})

	t.Run("short config flag", func(t *testing.T) {
		withArgs(t, "-c", "config.json")

		opts, err := ReadCliOptions()

		require.NoError(t, err)
		assert.Equal(t, "config.json", opts.ConfigPath)
	})
}

func TestServerReadFileOptions(t *testing.T) {
	t.Run("valid file", func(t *testing.T) {
		path := writeConfig(t, `{
			"address": "localhost:9090",
			"profile": true,
			"database_dsn": "postgres://localhost:5432/metrics",
			"hashing_key": "secret",
			"crypto_key": "private_key.pem",
			"store_file": "dump.json",
			"store_interval": 5,
			"restore": true,
			"audit_file": "audit.json",
			"audit_url": "http://localhost:9091/audit"
		}`)

		opts, err := ReadFileOptions(path)

		require.NoError(t, err)
		assert.Equal(t, &Options{
			Address:      "localhost:9090",
			Profile:      boolPtr(true),
			DatabaseDSN:  "postgres://localhost:5432/metrics",
			HashingKey:   "secret",
			CryptoKey:    "private_key.pem",
			DumpFile:     "dump.json",
			DumpInterval: uint64Ptr(5),
			DumpRestore:  boolPtr(true),
			AuditFile:    "audit.json",
			AuditURL:     "http://localhost:9091/audit",
		}, opts)
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := ReadFileOptions(filepath.Join(t.TempDir(), "missing.json"))

		assert.Error(t, err)
	})

	t.Run("invalid json", func(t *testing.T) {
		path := writeConfig(t, "{invalid")

		_, err := ReadFileOptions(path)

		assert.Error(t, err)
	})
}

func TestServerReadOptions(t *testing.T) {
	envKeys := []string{
		"ADDRESS", "PROFILE", "DATABASE_DSN",
		"KEY", "CRYPTO_KEY", "FILE_STORAGE_PATH",
		"STORE_INTERVAL", "RESTORE", "AUDIT_FILE", "AUDIT_URL", "CONFIG",
	}

	t.Run("env overrides cli", func(t *testing.T) {
		withArgs(t, "-a", "cli:1111", "-r")
		clearEnv(t, envKeys...)
		t.Setenv("ADDRESS", "env:2222")

		opts, err := ReadOptions()

		require.NoError(t, err)
		assert.Equal(t, "env:2222", opts.Address)
		require.NotNil(t, opts.DumpRestore)
		assert.True(t, *opts.DumpRestore)
	})

	t.Run("file used when config path set", func(t *testing.T) {
		path := writeConfig(t, `{"address": "file:3333", "store_interval": 30, "restore": true}`)

		withArgs(t, "-c", path)
		clearEnv(t, envKeys...)

		opts, err := ReadOptions()

		require.NoError(t, err)
		assert.Equal(t, "file:3333", opts.Address)
		require.NotNil(t, opts.DumpInterval)
		assert.Equal(t, uint64(30), *opts.DumpInterval)
		require.NotNil(t, opts.DumpRestore)
		assert.True(t, *opts.DumpRestore)
		assert.Equal(t, path, opts.ConfigPath)
	})

	t.Run("env and cli override file", func(t *testing.T) {
		path := writeConfig(t, `{"address": "file:3333", "store_interval": 30}`)

		withArgs(t, "-c", path, "-a", "cli:1111")
		clearEnv(t, envKeys...)
		t.Setenv("STORE_INTERVAL", "45")

		opts, err := ReadOptions()

		require.NoError(t, err)
		assert.Equal(t, "cli:1111", opts.Address)
		require.NotNil(t, opts.DumpInterval)
		assert.Equal(t, uint64(45), *opts.DumpInterval)
	})

	t.Run("missing config file", func(t *testing.T) {
		withArgs(t, "-config", filepath.Join(t.TempDir(), "missing.json"))
		clearEnv(t, envKeys...)

		_, err := ReadOptions()

		assert.Error(t, err)
	})
}
