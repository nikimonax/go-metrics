package config_test

import (
	"testing"
	"time"

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

	t.Run("invalid server type", func(t *testing.T) {
		tests := []config.ServerType{
			"",
			"tcp",
			"HTTP",
			"http ",
		}

		for _, serverType := range tests {
			cfg := newValidConfig()
			cfg.ServerType = serverType

			err := cfg.Validate()

			require.Error(t, err, "server type %q", serverType)
			assert.ErrorIs(t, err, config.ErrInvalidConfig)
			assert.ErrorContains(t, err, "invalid 'ServerType'")
		}
	})

	t.Run("empty listen", func(t *testing.T) {
		cfg := newValidConfig()
		cfg.Listen = ""

		err := cfg.Validate()

		require.Error(t, err)
		assert.ErrorContains(t, err, "required 'Listen'")
	})

	t.Run("database dsn", func(t *testing.T) {
		t.Run("empty dsn is allowed", func(t *testing.T) {
			cfg := newValidConfig()
			cfg.Database.DSN = ""

			require.NoError(t, cfg.Validate())
		})

		t.Run("valid dsn", func(t *testing.T) {
			tests := []string{
				"postgres://user:pass@localhost:5432/test",
				"postgres://user:pass@localhost:5432/test?sslmode=disable",
				// dsn without a scheme is allowed
				"localhost/test",
			}

			for _, dsn := range tests {
				cfg := newValidConfig()
				cfg.Database.DSN = dsn

				require.NoError(t, cfg.Validate(), "dsn %q", dsn)
			}
		})

		t.Run("invalid dsn", func(t *testing.T) {
			tests := []struct {
				dsn        string
				errMessage string
			}{
				// scheme other than postgres
				{"mysql://user:pass@localhost:3306/test", "unsupported database: mysql"},
				{"redis://localhost:6379/0", "unsupported database: redis"},
				// unparseable dsn
				{"postgres://user:pass@localhost:port/test", "invalid database dsn"},
			}

			for _, test := range tests {
				cfg := newValidConfig()
				cfg.Database.DSN = test.dsn

				err := cfg.Validate()

				require.Error(t, err, "dsn %q", test.dsn)
				assert.ErrorIs(t, err, config.ErrInvalidConfig)
				assert.ErrorContains(t, err, test.errMessage)
			}
		})
	})

	t.Run("audit url", func(t *testing.T) {
		t.Run("empty url is allowed", func(t *testing.T) {
			cfg := newValidConfig()
			cfg.Audit.URL = ""

			require.NoError(t, cfg.Validate())
		})

		t.Run("valid url", func(t *testing.T) {
			tests := []string{
				"http://127.0.0.1:8080/audit",
				"https://example.com/audit",
			}

			for _, url := range tests {
				cfg := newValidConfig()
				cfg.Audit.URL = url

				require.NoError(t, cfg.Validate(), "url %q", url)
			}
		})

		t.Run("invalid url", func(t *testing.T) {
			tests := []struct {
				url        string
				errMessage string
			}{
				// scheme other than http/https
				{"ftp://127.0.0.1:8080/audit", "unsupported audit url scheme: ftp"},
				{"tcp://127.0.0.1:8080/audit", "unsupported audit url scheme: tcp"},
				// unparseable url
				{"://invalid-url", "invalid audit url"},
			}

			for _, test := range tests {
				cfg := newValidConfig()
				cfg.Audit.URL = test.url

				err := cfg.Validate()

				require.Error(t, err, "url %q", test.url)
				assert.ErrorIs(t, err, config.ErrInvalidConfig)
				assert.ErrorContains(t, err, test.errMessage)
			}
		})
	})

	t.Run("dump config", func(t *testing.T) {
		t.Run("invalid dump interval", func(t *testing.T) {
			cfg := newValidConfig()
			cfg.Dump.Interval = -time.Second

			err := cfg.Validate()

			require.Error(t, err)
			assert.ErrorIs(t, err, config.ErrInvalidConfig)
			assert.ErrorContains(t, err, "required non negative 'Interval'")
		})
	})
}
