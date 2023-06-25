package modinfo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
)

func run(pass *analysis.Pass) (any, error) {
	return GetModuleInfo(pass)
}

func TestAnalyzer(t *testing.T) {
	// NOTE: analysistest does not yet support modules;
	// see https://github.com/golang/go/issues/37054 for details.
	// The workspaces are also not really supported, we can't run the analyzer at the root of the workspace.
	testCases := []struct {
		desc     string
		dir      string
		patterns []string
		len      int
	}{
		{
			desc:     "simple",
			dir:      "a",
			patterns: []string{"a"},
			len:      1,
		},
		{
			desc:     "module inside a workspace",
			dir:      "workspace",
			patterns: []string{"workspace/hello/..."},
			len:      2,
		},
		{
			desc:     "modules inside a workspace",
			dir:      "workspace",
			patterns: []string{"workspace/hello/...", "workspace/world/..."},
			len:      2,
		},
		{
			desc:     "bad module design",
			dir:      "badmodule",
			patterns: []string{"badmodule"},
			len:      1,
		},
	}

	a := Analyzer
	a.Run = run

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			results := analysistest.RunWithSuggestedFixes(t, analysistest.TestData(), a, test.patterns...)
			for _, result := range results {
				infos, ok := result.Result.([]ModInfo)
				require.True(t, ok)
				assert.Len(t, infos, test.len)
			}
		})
	}
}
