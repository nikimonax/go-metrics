package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/go-metrics/internal/server/config"
)

func TestServerConfigValidate(t *testing.T) {
	newValidConfig := func() config.ServerConfig {
		return config.NewDefaultConfig()
	}

	t.Run("default config is valid", func(t *testing.T) {
		require.NoError(t, newValidConfig().Validate())
	})

	t.Run("empty trusted subnet is allowed", func(t *testing.T) {
		cfg := newValidConfig()
		cfg.TrustedSubnet = ""

		require.NoError(t, cfg.Validate())
	})

	t.Run("valid trusted subnet", func(t *testing.T) {
		tests := []string{
			"192.168.0.0/24",
			"10.0.0.0/8",
			"0.0.0.0/0",
			"fd00::/8",
			"::/0",
		}

		for _, trustedSubnet := range tests {
			cfg := newValidConfig()
			cfg.TrustedSubnet = trustedSubnet

			require.NoError(t, cfg.Validate(), "trusted subnet %q", trustedSubnet)
		}
	})

	t.Run("invalid trusted subnet", func(t *testing.T) {
		tests := []string{
			"not-a-subnet",
			// bare ip without prefix is not a valid subnet
			"192.168.0.1",
			"192.168.0.0/33",
		}

		for _, trustedSubnet := range tests {
			cfg := newValidConfig()
			cfg.TrustedSubnet = trustedSubnet

			err := cfg.Validate()

			require.Error(t, err, "trusted subnet %q", trustedSubnet)
			assert.ErrorIs(t, err, config.ErrInvalidConfig)
			assert.ErrorContains(t, err, "invalid 'TrustedSubnet'")
		}
	})

	t.Run("trusted subnet error joined with other errors", func(t *testing.T) {
		cfg := newValidConfig()
		cfg.TrustedSubnet = "not-a-subnet"
		cfg.Dump.File = ""

		err := cfg.Validate()

		require.Error(t, err)
		assert.ErrorContains(t, err, "required 'File'")
	})
}
