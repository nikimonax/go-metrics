package structreset

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/otiai10/copy"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis/analysistest"
	"golang.org/x/tools/go/packages"
)

func copyTestdata(t *testing.T, dst string) {
	t.Helper()

	err := copy.Copy(analysistest.TestData(), dst)
	require.NoError(t, err)
}

func TestGenerate(t *testing.T) {
	root := t.TempDir()
	copyTestdata(t, root)

	cfg := packages.Config{
		Mode: packages.LoadAllSyntax,
		Dir:  root,
	}

	pkgs, err := packages.Load(&cfg, "./pkg1", "./pkg2")
	require.NoError(t, err)

	generated, err := Generate(pkgs)
	require.NoError(t, err)
	require.Len(t, generated, 2) // pkg1 & pkg2

	for pkg, file := range generated {
		filePath := filepath.Join(pkg.Dir, GeneratedFileName)
		require.NoError(t, writeFile(file, filePath))
	}

	runE2ETests(t, root)
}

func runE2ETests(t *testing.T, root string) {
	t.Helper()

	cmd := exec.CommandContext(t.Context(), "go", "test", "./...")
	cmd.Dir = root

	out, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("go test failed:\n%s", out)
	}
}
