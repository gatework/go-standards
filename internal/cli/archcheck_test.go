package cli_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gatework/go-standards/internal/cli"
)

func fakeList(packages string) cli.Lister {
	return func(_ context.Context, _ string, args ...string) ([]byte, error) {
		if len(args) == 1 && args[0] == "-m" {
			return []byte("example.com/m\n"), nil
		}
		return []byte(packages), nil
	}
}

func TestArchcheck(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	rules := filepath.Join(dir, "architecture.yaml")
	body := "declared: [., ./domain]\nrules:\n  - name: domain stays pure\n    packages: [./domain]\n    deny: [os]\n"
	if err := os.WriteFile(rules, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	clean := `{"ImportPath":"example.com/m"}{"ImportPath":"example.com/m/domain","Imports":["strings"]}`
	dirty := `{"ImportPath":"example.com/m"}{"ImportPath":"example.com/m/domain","Imports":["os"]}`
	for name, tc := range map[string]struct {
		packages string
		code     int
		stderr   string
	}{
		"pure":     {clean, cli.ExitOK, ""},
		"violates": {dirty, cli.ExitDeviation, "domain stays pure"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var stdout, stderr bytes.Buffer
			code := cli.Archcheck(t.Context(), []string{"-dir", dir}, fakeList(tc.packages), &stdout, &stderr)
			if code != tc.code || !strings.Contains(stderr.String(), tc.stderr) {
				t.Fatalf("exit %d stderr %q", code, stderr.String())
			}
		})
	}
	var stderr bytes.Buffer
	if code := cli.Archcheck(t.Context(), []string{"-dir", dir, "-rules", "missing.yaml"}, fakeList(clean), &bytes.Buffer{}, &stderr); code != cli.ExitUsage {
		t.Fatalf("missing rules: exit %d", code)
	}
}
