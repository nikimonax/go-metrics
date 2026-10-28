package staticlint

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuild(t *testing.T) {
	analyzers := Build()
	assert.LessOrEqual(t, len(analyzers), maxCap)
}
