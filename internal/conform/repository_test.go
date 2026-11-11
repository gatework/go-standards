package conform_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	standards "github.com/gatework/go-standards"
	"github.com/gatework/go-standards/internal/conform"
)

func repo(t *testing.T, files map[string]string) *conform.Repository {
	t.Helper()
	dir := t.TempDir()
	files["go.mod"] = "module example.com/x\n"
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	r, err := conform.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func kinds(t *testing.T, r *conform.Repository) []conform.Kind {
	t.Helper()
	ds, err := r.Check()
	if err != nil {
		t.Fatal(err)
	}
	out := []conform.Kind{}
	for _, d := range ds {
		out = append(out, d.Kind)
	}
	return out
}

func TestRepository_Check(t *testing.T) {
	t.Parallel()
	agents := "# x\n\nFollow [the standard](../go-standards/STANDARD.md).\n"
	mk := string(standards.MakeTargets())
	for name, tc := range map[string]struct {
		files map[string]string
		want  []conform.Kind
	}{
		"conforming":        {map[string]string{conform.LintFile: string(standards.LintConfig()), conform.MakeFile: mk, conform.AgentsFile: agents}, []conform.Kind{}},
		"missing all":       {map[string]string{}, []conform.Kind{conform.KindMissing, conform.KindMissing, conform.KindMissing}},
		"edited copy":       {map[string]string{conform.LintFile: string(standards.LintConfig()) + "# local\n", conform.MakeFile: mk, conform.AgentsFile: agents}, []conform.Kind{conform.KindModified}},
		"older version":     {map[string]string{conform.LintFile: "# standard-version: 0.9.0\n", conform.MakeFile: mk, conform.AgentsFile: agents}, []conform.Kind{conform.KindOutdated}},
		"unlinked agents":   {map[string]string{conform.LintFile: string(standards.LintConfig()), conform.MakeFile: mk, conform.AgentsFile: "# x\n"}, []conform.Kind{conform.KindNoMarker}},
		"marker too far in": {map[string]string{conform.LintFile: string(standards.LintConfig()), conform.MakeFile: mk, conform.AgentsFile: "1\n2\n3\n4\n5\n6 go-standards/STANDARD.md\n"}, []conform.Kind{conform.KindNoMarker}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := kinds(t, repo(t, tc.files)); !slices.Equal(got, tc.want) {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestRepository_Write(t *testing.T) {
	t.Parallel()
	r := repo(t, map[string]string{conform.LintFile: "stale\n", conform.AgentsFile: "see go-standards/STANDARD.md\n"})
	if err := r.Write(); err != nil {
		t.Fatal(err)
	}
	if got := kinds(t, r); len(got) != 0 {
		t.Fatalf("still deviates after write: %v", got)
	}
}

func TestOpen_RejectsNonModule(t *testing.T) {
	t.Parallel()
	if _, err := conform.Open(t.TempDir()); err == nil {
		t.Fatal("opened a directory without go.mod")
	}
}
