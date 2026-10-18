package server

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nikimonax/go-metrics/internal/server/config"
)

func TestServer(t *testing.T) {
	cfg := config.NewDefaultConfig()

	server, err := New(&cfg)
	require.NoError(t, err)
	require.NoError(t, server.app.Err())

	err = server.app.Start(t.Context())
	require.NoError(t, err)

	err = server.app.Stop(t.Context())
	require.NoError(t, err)
}
