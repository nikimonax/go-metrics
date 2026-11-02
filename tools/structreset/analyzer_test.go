package structreset_test

import (
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/stretchr/testify/require"

	"github.com/nikimonax/go-metrics/tools/structreset"
)

func TestStructResetAnalyzer(t *testing.T) {
	alanyzers := []*analysis.Analyzer{structreset.Analyzer}
	require.NoError(t, analysis.Validate(alanyzers))

	analysistest.Run(t, analysistest.TestData(), structreset.Analyzer, "./...")
}
