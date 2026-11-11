package standards_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	standards "github.com/gatework/go-standards"
)

// TestCheckRunsExtraSetAfterInclude guards the documented contract that a
// repository may set CHECK_EXTRA after including standard.mk.
func TestCheckRunsExtraSetAfterInclude(t *testing.T) {
	t.Parallel()
	makePath, err := exec.LookPath("make")
	if err != nil {
		t.Skip("make not installed")
	}
	dir := t.TempDir()
	makefile := "include standard.mk\nCHECK_EXTRA := extra\n.PHONY: extra\nextra:\n\t@echo EXTRA-RAN\n"
	for name, body := range map[string][]byte{"standard.mk": standards.MakeTargets(), "Makefile": []byte(makefile)} {
		if err := os.WriteFile(filepath.Join(dir, name), body, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	// -n prints the shared recipes without running them; the recursive make
	// for CHECK_EXTRA still runs and prints the extra recipe.
	cmd := exec.CommandContext(t.Context(), makePath, "-n", "check", "GO=true", "STDSYNC=true", "ARCHCHECK=true", "GOLANGCI_LINT=true", "GOVULNCHECK=true")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("make: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "echo EXTRA-RAN") {
		t.Fatalf("check did not run CHECK_EXTRA:\n%s", out)
	}
}
