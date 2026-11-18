package server_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/zap/zapcore"

	"github.com/nikimonax/go-metrics/internal/server"
	"github.com/nikimonax/go-metrics/internal/server/config"
)

func writeArmoredPrivateKey(t *testing.T) string {
	t.Helper()

	key, err := crypto.PGP().
		KeyGeneration().
		AddUserId("Test", "test@example.com").
		New().
		GenerateKey()

	require.NoError(t, err)

	armored, err := key.Armor()
	require.NoError(t, err)

	path := filepath.Join(t.TempDir(), "private.key")
	require.NoError(t, os.WriteFile(path, []byte(armored), 0o600))

	return path
}

func TestServer(t *testing.T) {
	keyFile := writeArmoredPrivateKey(t)

	tests := []struct {
		name  string
		setup func(cfg *config.ServerConfig)
	}{
		{
			name:  "default",
			setup: nil,
		},
		{
			name: "profile",
			setup: func(cfg *config.ServerConfig) {
				cfg.Profile = true
			},
		},
		{
			name: "database",
			setup: func(cfg *config.ServerConfig) {
				cfg.Database.DSN = "postgres://user:pass@localhost:5432/test?sslmode=disable"
				cfg.Database.Migrate = false
			},
		},
		{
			name: "profile and database",
			setup: func(cfg *config.ServerConfig) {
				cfg.Profile = true
				cfg.Database.DSN = "postgres://user:pass@localhost:5432/test?sslmode=disable"
				cfg.Database.Migrate = false
			},
		},
		{
			name: "grpc",
			setup: func(cfg *config.ServerConfig) {
				cfg.ServerType = config.GRPCServer
			},
		},
		{
			name: "dump restore",
			setup: func(cfg *config.ServerConfig) {
				cfg.Dump.Restore = true
				cfg.Dump.File = filepath.Join(t.TempDir(), "metrics.json")
			},
		},
		{
			name: "audit file",
			setup: func(cfg *config.ServerConfig) {
				cfg.Audit.File = filepath.Join(t.TempDir(), "audit.log")
			},
		},
		{
			name: "audit url",
			setup: func(cfg *config.ServerConfig) {
				cfg.Audit.URL = "http://127.0.0.1:1/audit"
			},
		},
		{
			name: "hashing key",
			setup: func(cfg *config.ServerConfig) {
				cfg.Security.HashingKey = "test-secret"
			},
		},
		{
			name: "crypto key",
			setup: func(cfg *config.ServerConfig) {
				cfg.Security.CryptoKeyFile = keyFile
			},
		},
		{
			name: "all features",
			setup: func(cfg *config.ServerConfig) {
				cfg.Profile = true
				cfg.TrustedSubnet = "192.168.0.0/24"
				cfg.Database.DSN = "postgres://user:pass@localhost:5432/test?sslmode=disable"
				cfg.Database.Migrate = false
				cfg.Security.HashingKey = "test-secret"
				cfg.Security.CryptoKeyFile = keyFile
				cfg.Dump.Restore = true
				cfg.Dump.File = filepath.Join(t.TempDir(), "metrics.json")
				cfg.Audit.File = filepath.Join(t.TempDir(), "audit.log")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := config.NewDefaultConfig()
			cfg.Log.Level = zapcore.WarnLevel

			if test.setup != nil {
				test.setup(&cfg)
			}

			app := fx.New(server.BuildOpts(&cfg)...)

			require.NoError(t, app.Err())
			require.NoError(t, app.Start(t.Context()))
			require.NoError(t, app.Stop(t.Context()))
		})
	}
}
