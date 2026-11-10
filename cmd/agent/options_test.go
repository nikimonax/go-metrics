package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/go-metrics/internal/agent/config"
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

func TestToAgentConfig(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		wantErr bool
		check   func(t *testing.T, cfg *config.AgentConfig)
	}{
		{
			name: "empty options: defaults kept",
			check: func(t *testing.T, cfg *config.AgentConfig) {
				want := config.NewDefaultConfig()

				assert.Equal(t, want.BaseURL, cfg.BaseURL)
				assert.Equal(t, want.APIVersion, cfg.APIVersion)
				assert.Zero(t, cfg.RateLimit)
				assert.Equal(t, want.PollInterval, cfg.PollInterval)
				assert.Equal(t, want.ReportInterval, cfg.ReportInterval)
				assert.Empty(t, cfg.Security.HashingKey)
				assert.Empty(t, cfg.Security.CryptoKeyFile)
			},
		},
		{
			name: "address without scheme",
			opts: Options{Address: "localhost:9090"},
			check: func(t *testing.T, cfg *config.AgentConfig) {
				assert.Equal(t, "http://localhost:9090", cfg.BaseURL.String())
			},
		},
		{
			name: "address with http scheme",
			opts: Options{Address: "http://localhost:9090"},
			check: func(t *testing.T, cfg *config.AgentConfig) {
				assert.Equal(t, "http://localhost:9090", cfg.BaseURL.String())
			},
		},
		{
			name: "address with https scheme",
			opts: Options{Address: "https://localhost:9090"},
			check: func(t *testing.T, cfg *config.AgentConfig) {
				assert.Equal(t, "https://localhost:9090", cfg.BaseURL.String())
			},
		},
		{
			name:    "invalid address",
			opts:    Options{Address: "http://exa mple.com"},
			wantErr: true,
		},
		{
			name: "all fields set",
			opts: Options{
				Address:            "localhost:9090",
				APIVersion:         3,
				RateLimit:          5,
				PollIntervalSecs:   2,
				ReportIntervalSecs: 10,
				HashingKey:         "secret",
				CryptoKey:          "public_key.pem",
			},
			check: func(t *testing.T, cfg *config.AgentConfig) {
				assert.Equal(t, "http://localhost:9090", cfg.BaseURL.String())
				assert.Equal(t, uint(3), cfg.APIVersion)
				assert.Equal(t, int64(5), cfg.RateLimit)
				assert.Equal(t, 2*time.Second, cfg.PollInterval)
				assert.Equal(t, 10*time.Second, cfg.ReportInterval)
				assert.Equal(t, "secret", cfg.Security.HashingKey)
				assert.Equal(t, "public_key.pem", cfg.Security.CryptoKeyFile)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := tc.opts.ToAgentConfig()

			if tc.wantErr {
				require.Error(t, err)
				assert.ErrorContains(t, err, "failed parse url")
				assert.Nil(t, cfg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)
			tc.check(t, cfg)
		})
	}
}

func TestAgentOptionsMerge(t *testing.T) {
	tests := []struct {
		name  string
		base  Options
		other Options
		want  Options
	}{
		{
			name: "empty other keeps base",
			base: Options{
				Address:            "base:1111",
				APIVersion:         1,
				RateLimit:          2,
				PollIntervalSecs:   3,
				ReportIntervalSecs: 4,
				HashingKey:         "key",
				CryptoKey:          "crypto",
				ConfigPath:         "cfg.json",
			},
			other: Options{},
			want: Options{
				Address:            "base:1111",
				APIVersion:         1,
				RateLimit:          2,
				PollIntervalSecs:   3,
				ReportIntervalSecs: 4,
				HashingKey:         "key",
				CryptoKey:          "crypto",
				ConfigPath:         "cfg.json",
			},
		},
		{
			name: "non-empty other overrides",
			base: Options{
				Address:    "base:1111",
				APIVersion: 1,
			},
			other: Options{
				Address:            "other:2222",
				APIVersion:         2,
				RateLimit:          3,
				PollIntervalSecs:   4,
				ReportIntervalSecs: 5,
				HashingKey:         "key",
				CryptoKey:          "crypto",
				ConfigPath:         "cfg.json",
			},
			want: Options{
				Address:            "other:2222",
				APIVersion:         2,
				RateLimit:          3,
				PollIntervalSecs:   4,
				ReportIntervalSecs: 5,
				HashingKey:         "key",
				CryptoKey:          "crypto",
				ConfigPath:         "cfg.json",
			},
		},
		{
			name:  "zero values do not override",
			base:  Options{Address: "base:1111", APIVersion: 1},
			other: Options{RateLimit: 3},
			want:  Options{Address: "base:1111", APIVersion: 1, RateLimit: 3},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			opts := tc.base
			err := opts.Merge(&tc.other)

			require.NoError(t, err)
			assert.Equal(t, tc.want, opts)
		})
	}
}

func TestAgentReadEnvOptions(t *testing.T) {
	t.Run("empty environment", func(t *testing.T) {
		clearEnv(t,
			"ADDRESS", "API", "RATE_LIMIT",
			"POLL_INTERVAL", "REPORT_INTERVAL",
			"KEY", "CRYPTO_KEY", "CONFIG",
		)

		opts, err := ReadEnvOptions()

		require.NoError(t, err)
		assert.Equal(t, &Options{}, opts)
	})

	t.Run("all variables set", func(t *testing.T) {
		t.Setenv("ADDRESS", "localhost:9090")
		t.Setenv("API", "3")
		t.Setenv("RATE_LIMIT", "5")
		t.Setenv("POLL_INTERVAL", "2")
		t.Setenv("REPORT_INTERVAL", "10")
		t.Setenv("KEY", "secret")
		t.Setenv("CRYPTO_KEY", "public_key.pem")

		opts, err := ReadEnvOptions()

		require.NoError(t, err)
		assert.Equal(t, &Options{
			Address:            "localhost:9090",
			APIVersion:         3,
			RateLimit:          5,
			PollIntervalSecs:   2,
			ReportIntervalSecs: 10,
			HashingKey:         "secret",
			CryptoKey:          "public_key.pem",
		}, opts)
	})

	t.Run("invalid uint value", func(t *testing.T) {
		t.Setenv("API", "not-a-number")

		_, err := ReadEnvOptions()

		assert.Error(t, err)
	})
}

func TestAgentReadCliOptions(t *testing.T) {
	t.Run("no flags", func(t *testing.T) {
		withArgs(t)

		opts, err := ReadCliOptions()

		require.NoError(t, err)
		assert.Equal(t, &Options{}, opts)
	})

	t.Run("all flags", func(t *testing.T) {
		withArgs(t,
			"-a", "localhost:9090",
			"-v", "3",
			"-l", "5",
			"-p", "2",
			"-r", "10",
			"-k", "secret",
			"-crypto-key", "public_key.pem",
			"-config", "config.json",
		)

		opts, err := ReadCliOptions()

		require.NoError(t, err)
		assert.Equal(t, &Options{
			Address:            "localhost:9090",
			APIVersion:         3,
			RateLimit:          5,
			PollIntervalSecs:   2,
			ReportIntervalSecs: 10,
			HashingKey:         "secret",
			CryptoKey:          "public_key.pem",
			ConfigPath:         "config.json",
		}, opts)
	})

	t.Run("short config flag", func(t *testing.T) {
		withArgs(t, "-c", "config.json")

		opts, err := ReadCliOptions()

		require.NoError(t, err)
		assert.Equal(t, "config.json", opts.ConfigPath)
	})
}

func TestAgentReadFileOptions(t *testing.T) {
	t.Run("valid file", func(t *testing.T) {
		path := writeConfig(t, `{
			"address": "localhost:9090",
			"api_version": 3,
			"rate_limit": 5,
			"poll_interval": 2,
			"report_interval": 10,
			"hashing_key": "secret",
			"crypto_key": "public_key.pem"
		}`)

		opts, err := ReadFileOptions(path)

		require.NoError(t, err)
		assert.Equal(t, &Options{
			Address:            "localhost:9090",
			APIVersion:         3,
			RateLimit:          5,
			PollIntervalSecs:   2,
			ReportIntervalSecs: 10,
			HashingKey:         "secret",
			CryptoKey:          "public_key.pem",
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

func TestAgentReadOptions(t *testing.T) {
	envKeys := []string{
		"ADDRESS", "API", "RATE_LIMIT",
		"POLL_INTERVAL", "REPORT_INTERVAL",
		"KEY", "CRYPTO_KEY", "CONFIG",
	}

	t.Run("env overrides cli", func(t *testing.T) {
		withArgs(t, "-a", "cli:1111", "-p", "1")
		clearEnv(t, envKeys...)
		t.Setenv("ADDRESS", "env:2222")

		opts, err := ReadOptions()

		require.NoError(t, err)
		assert.Equal(t, "env:2222", opts.Address)
		assert.Equal(t, uint64(1), opts.PollIntervalSecs)
	})

	t.Run("file used when config path set", func(t *testing.T) {
		path := writeConfig(t, `{"address": "file:3333", "poll_interval": 30}`)

		withArgs(t, "-c", path)
		clearEnv(t, envKeys...)

		opts, err := ReadOptions()

		require.NoError(t, err)
		assert.Equal(t, "file:3333", opts.Address)
		assert.Equal(t, uint64(30), opts.PollIntervalSecs)
		assert.Equal(t, path, opts.ConfigPath)
	})

	t.Run("env and cli override file", func(t *testing.T) {
		path := writeConfig(t, `{"address": "file:3333", "poll_interval": 30}`)

		withArgs(t, "-c", path, "-a", "cli:1111")
		clearEnv(t, envKeys...)
		t.Setenv("POLL_INTERVAL", "45")

		opts, err := ReadOptions()

		require.NoError(t, err)
		assert.Equal(t, "cli:1111", opts.Address)
		assert.Equal(t, uint64(45), opts.PollIntervalSecs)
	})

	t.Run("missing config file", func(t *testing.T) {
		withArgs(t, "-config", filepath.Join(t.TempDir(), "missing.json"))
		clearEnv(t, envKeys...)

		_, err := ReadOptions()

		assert.Error(t, err)
	})
}
