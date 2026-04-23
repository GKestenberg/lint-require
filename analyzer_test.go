package lintrequire_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	lintrequire "github.com/GKestenberg/lint-require"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, lintrequire.Analyzer, "a", "c", "d")
}
