package netutil_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/go-metrics/internal/lib/netutil"
)

func TestParsePrefixes(t *testing.T) {
	prefixes, err := netutil.ParsePrefixes([]string{"not-an-ip", "192.168.0.0/24", "bad..entry"})

	require.Error(t, err)
	assert.Nil(t, prefixes)
	assert.ErrorContains(t, err, "invalid IP or CIDR: not-an-ip")
	assert.ErrorContains(t, err, "invalid IP or CIDR: bad..entry")

}
