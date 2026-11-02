package staticlint

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/analysis"
)

func TestBuild(t *testing.T) {
	analyzers := Build()

	assert.Equal(t, len(analyzers), maxCap)
	assert.NoError(t, analysis.Validate(analyzers))
}
