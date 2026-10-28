package exitcheck_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/nikimonax/go-metrics/tools/exitcheck"
)

func TestExitCheckAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), exitcheck.Analyzer, "./...")
}
