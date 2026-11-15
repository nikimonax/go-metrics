package netutil_test

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/go-metrics/internal/lib/netutil"
)

func TestGetSourceIPFor(t *testing.T) {
	t.Run("loopback target returns the target itself", func(t *testing.T) {
		tests := []string{"127.0.0.1", "::1"}

		for _, addr := range tests {
			target := net.ParseIP(addr)
			require.NotNil(t, target)

			sourceIP, err := netutil.GetSourceIPFor(target)

			require.NoError(t, err)
			assert.True(t, sourceIP.Equal(target), "target %s", addr)
		}
	})

	t.Run("target of a local interface returns the interface address", func(t *testing.T) {
		addrs, err := net.InterfaceAddrs()
		require.NoError(t, err)

		var target net.IP
		local := make(map[string]bool, len(addrs))

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)

			if !ok || ipNet.IP.IsLoopback() || ipNet.IP.IsLinkLocalUnicast() {
				continue
			}

			if target == nil {
				target = ipNet.IP
			}

			local[ipNet.IP.String()] = true
		}

		if target == nil {
			t.Skip("no non-loopback interface addresses available")
		}

		sourceIP, err := netutil.GetSourceIPFor(target)

		require.NoError(t, err)
		assert.True(
			t, local[sourceIP.String()],
			"source ip %s must be one of the local interface addresses", sourceIP,
		)
	})

	t.Run("external target returns error", func(t *testing.T) {
		target := net.ParseIP("203.0.113.1")
		require.NotNil(t, target)

		_, err := netutil.GetSourceIPFor(target)

		require.ErrorIs(t, err, netutil.ErrExternalTarget)
	})
}
