package agent

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/go-metrics/internal/agent/config"
)

func writeArmoredPublicKey(t *testing.T) string {
	t.Helper()

	key, err := crypto.PGP().
		KeyGeneration().
		AddUserId("Test", "test@example.com").
		New().
		GenerateKey()

	require.NoError(t, err)

	armored, err := key.GetArmoredPublicKey()
	require.NoError(t, err)

	path := filepath.Join(t.TempDir(), "public.key")
	require.NoError(t, os.WriteFile(path, []byte(armored), 0o600))

	return path
}

func TestAgent(t *testing.T) {
	keyFile := writeArmoredPublicKey(t)

	tests := []struct {
		name  string
		setup func(cfg *config.AgentConfig)
	}{
		{
			name:  "default",
			setup: nil,
		},
		{
			name: "grpc",
			setup: func(cfg *config.AgentConfig) {
				cfg.GrpcMode = true
			},
		},
		{
			name: "grpc with source ip",
			setup: func(cfg *config.AgentConfig) {
				cfg.GrpcMode = true
				cfg.SourceIP = "127.0.0.1"
			},
		},
		{
			name: "api v2",
			setup: func(cfg *config.AgentConfig) {
				cfg.APIVersion = 2
			},
		},
		{
			name: "api v3",
			setup: func(cfg *config.AgentConfig) {
				cfg.APIVersion = 3
			},
		},
		{
			name: "rate limit",
			setup: func(cfg *config.AgentConfig) {
				cfg.RateLimit = 10
			},
		},
		{
			name: "source ip",
			setup: func(cfg *config.AgentConfig) {
				cfg.SourceIP = "127.0.0.1"
			},
		},
		{
			name: "backoff disabled",
			setup: func(cfg *config.AgentConfig) {
				cfg.Backoff.Retry = 0
			},
		},
		{
			name: "backoff custom",
			setup: func(cfg *config.AgentConfig) {
				cfg.Backoff.Retry = 2
				cfg.Backoff.Seed = 10 * time.Millisecond
				cfg.Backoff.Add = 20 * time.Millisecond
				cfg.Backoff.Mul = 2
			},
		},
		{
			name: "hashing key",
			setup: func(cfg *config.AgentConfig) {
				cfg.APIVersion = 2
				cfg.Security.HashingKey = "test-secret"
			},
		},
		{
			name: "crypto key",
			setup: func(cfg *config.AgentConfig) {
				cfg.APIVersion = 3
				cfg.Security.CryptoKeyFile = keyFile
			},
		},
		{
			name: "all features",
			setup: func(cfg *config.AgentConfig) {
				cfg.APIVersion = 3
				cfg.RateLimit = 10
				cfg.SourceIP = "127.0.0.1"
				cfg.Security.HashingKey = "test-secret"
				cfg.Security.CryptoKeyFile = keyFile
				cfg.Backoff.Retry = 2
				cfg.Backoff.Seed = 10 * time.Millisecond
				cfg.Backoff.Add = 20 * time.Millisecond
				cfg.Backoff.Mul = 2
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := config.NewDefaultConfig()

			if test.setup != nil {
				test.setup(&cfg)
			}

			agent, err := New(&cfg)

			require.NoError(t, err)
			require.NoError(t, agent.app.Err())
			require.NoError(t, agent.app.Start(t.Context()))
			require.NoError(t, agent.app.Stop(t.Context()))
		})
	}
}
