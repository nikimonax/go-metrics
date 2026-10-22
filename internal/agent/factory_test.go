package agent

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nikimonax/go-metrics/internal/agent/config"
)

func TestServer(t *testing.T) {
	cfg := config.NewDefaultConfig()

	agent, err := New(&cfg)
	require.NoError(t, err)
	require.NoError(t, agent.app.Err())

	err = agent.app.Start(t.Context())
	require.NoError(t, err)

	err = agent.app.Stop(t.Context())
	require.NoError(t, err)
}
