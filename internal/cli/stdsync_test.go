package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	standards "github.com/gatework/go-standards"
	"github.com/gatework/go-standards/internal/cli"
)

func TestStdsync(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("see go-standards/STANDARD.md\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	steps := []struct {
		args []string
		code int
		out  string
	}{
		{[]string{"check", dir}, cli.ExitDeviation, ".golangci.yml: missing"},
		{[]string{"write", dir}, cli.ExitOK, "wrote .golangci.yml"},
		{[]string{"check", dir}, cli.ExitOK, ""},
		{[]string{"version"}, cli.ExitOK, standards.Version()},
		{[]string{"bogus"}, cli.ExitUsage, ""},
		{nil, cli.ExitUsage, ""},
	}
	for _, s := range steps {
		var stdout, stderr bytes.Buffer
		if code := cli.Stdsync(s.args, &stdout, &stderr); code != s.code {
			t.Fatalf("%v: exit %d want %d (%s)", s.args, code, s.code, stderr.String())
		}
		if !strings.Contains(stdout.String(), s.out) {
			t.Fatalf("%v: output %q lacks %q", s.args, stdout.String(), s.out)
		}
	}
}
