package config_test

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/go-metrics/internal/agent/config"
	"github.com/nikimonax/go-metrics/internal/impl/gateway"
)

func parseBaseURL(t *testing.T, rawURL string) *url.URL {
	t.Helper()

	baseURL, err := url.Parse(rawURL)
	require.NoError(t, err)

	return baseURL
}

func TestAgentConfigGetSourceIPWithFallback(t *testing.T) {
	t.Run("explicit source ip is returned as is", func(t *testing.T) {
		cfg := config.NewDefaultConfig()
		cfg.SourceIP = "192.0.2.1"

		sourceIP, err := cfg.GetSourceIPWithFallback(t.Context())

		require.NoError(t, err)
		assert.Equal(t, "192.0.2.1", sourceIP)
	})

	t.Run("loopback target", func(t *testing.T) {
		tests := []struct {
			name    string
			baseURL string
			want    string
		}{
			{
				name:    "ipv4",
				baseURL: "http://127.0.0.1:9090",
				want:    "127.0.0.1",
			},
			{
				name:    "ipv6",
				baseURL: "http://[::1]:9090",
				want:    "::1",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				cfg := config.NewDefaultConfig()
				cfg.BaseURL = parseBaseURL(t, tc.baseURL)

				sourceIP, err := cfg.GetSourceIPWithFallback(t.Context())

				require.NoError(t, err)
				assert.Equal(t, tc.want, sourceIP)
			})
		}
	})

	t.Run("external target returns empty source ip", func(t *testing.T) {
		cfg := config.NewDefaultConfig()
		cfg.BaseURL = parseBaseURL(t, "http://203.0.113.1:9090")

		sourceIP, err := cfg.GetSourceIPWithFallback(t.Context())

		require.NoError(t, err)
		assert.Empty(t, sourceIP)
	})

	t.Run("lookup failure returns error", func(t *testing.T) {
		cfg := config.NewDefaultConfig()
		cfg.BaseURL = parseBaseURL(t, "http://unreachable.invalid:9090")

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		_, err := cfg.GetSourceIPWithFallback(ctx)

		require.Error(t, err)
	})
}

func TestAgentConfigValidate(t *testing.T) {
	t.Run("default config is valid", func(t *testing.T) {
		require.NoError(t, config.NewDefaultConfig().Validate())
	})

	t.Run("empty source ip is allowed", func(t *testing.T) {
		cfg := config.NewDefaultConfig()
		cfg.SourceIP = ""

		require.NoError(t, cfg.Validate())
	})

	t.Run("valid source ip", func(t *testing.T) {
		tests := []string{"192.0.2.1", "10.0.0.1", "::1", "fd00::1"}

		for _, sourceIP := range tests {
			cfg := config.NewDefaultConfig()
			cfg.SourceIP = sourceIP

			require.NoError(t, cfg.Validate(), "source ip %q", sourceIP)
		}
	})

	t.Run("invalid source ip", func(t *testing.T) {
		cfg := config.NewDefaultConfig()
		cfg.SourceIP = "not-an-ip"

		err := cfg.Validate()

		require.Error(t, err)
		assert.ErrorIs(t, err, config.ErrInvalidConfig)
		assert.ErrorContains(t, err, "invalid 'SourceIP' address")
	})

	t.Run("nil base url", func(t *testing.T) {
		cfg := config.NewDefaultConfig()
		cfg.BaseURL = nil

		err := cfg.Validate()

		require.Error(t, err)
		assert.ErrorContains(t, err, "required non-nil 'BaseURL'")
	})

	t.Run("joins all validation errors", func(t *testing.T) {
		cfg := config.AgentConfig{
			SourceIP:       "not-an-ip",
			RateLimit:      -1,
			PollInterval:   -1,
			ReportInterval: -1,
		}

		err := cfg.Validate()

		require.Error(t, err)
		assert.ErrorIs(t, err, config.ErrInvalidConfig)

		for _, msg := range []string{
			"invalid 'SourceIP' address",
			fmt.Sprintf(
				"required 'APIVersion' equals 1-%d",
				gateway.GetMaxAPIVersion(),
			),
			"required 'RateLimit' greater or equals 0",
			"required 'PollInterval' greater than 0",
			"required 'ReportInterval' greater than 0",
		} {
			assert.ErrorContains(t, err, msg)
		}
	})
}
