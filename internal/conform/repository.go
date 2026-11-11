package conform

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	standards "github.com/gatework/go-standards"
)

// File names a repository must carry.
const (
	LintFile   = ".golangci.yml"
	MakeFile   = "standard.mk"
	AgentsFile = "AGENTS.md"
)

// managed are the distributed files: each copy must equal its original.
var managed = []struct {
	name     string
	original func() []byte
}{
	{LintFile, standards.LintConfig},
	{MakeFile, standards.MakeTargets},
}

// AgentsMarker is the reference every AGENTS.md must contain in its first
// lines so that agents load the standard before repository-specific rules.
const AgentsMarker = "go-standards/STANDARD.md"

// agentsHeadLines is how far into AGENTS.md the marker may appear.
const agentsHeadLines = 5

// Kind classifies a deviation.
type Kind string

// Deviation kinds.
const (
	KindMissing  Kind = "missing"
	KindModified Kind = "modified"
	KindOutdated Kind = "outdated"
	KindNoMarker Kind = "no_standard_reference"
)

// Deviation is one way a repository differs from the standard.
type Deviation struct {
	File   string
	Kind   Kind
	Detail string
}

func (d Deviation) String() string {
	if d.Detail == "" {
		return fmt.Sprintf("%s: %s", d.File, d.Kind)
	}
	return fmt.Sprintf("%s: %s (%s)", d.File, d.Kind, d.Detail)
}

// Repository is a checked-out module root.
type Repository struct {
	dir string
}

// Open returns the repository rooted at dir, which must contain go.mod.
func Open(dir string) (*Repository, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("conform: resolve %s: %w", dir, err)
	}
	if _, err := os.Stat(filepath.Join(abs, "go.mod")); err != nil {
		return nil, fmt.Errorf("conform: %s is not a module root: %w", abs, err)
	}
	return &Repository{dir: abs}, nil
}

// Dir returns the repository root.
func (r *Repository) Dir() string { return r.dir }

// Check lists every deviation from the standard; none means conforming.
func (r *Repository) Check() ([]Deviation, error) {
	var out []Deviation
	for _, m := range managed {
		d, deviates, err := r.compare(m.name, m.original())
		if err != nil {
			return nil, err
		}
		if deviates {
			out = append(out, d)
		}
	}
	agents, err := r.read(AgentsFile)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		out = append(out, Deviation{File: AgentsFile, Kind: KindMissing})
	case err != nil:
		return nil, err
	case !referencesStandard(agents):
		out = append(out, Deviation{File: AgentsFile, Kind: KindNoMarker, Detail: "first lines must link " + AgentsMarker})
	}
	return out, nil
}

// Write replaces every distributed file with its original. AGENTS.md is
// repository-owned and never rewritten.
func (r *Repository) Write() error {
	for _, m := range managed {
		path := filepath.Join(r.dir, m.name)
		if err := os.WriteFile(path, m.original(), 0o644); err != nil { //nolint:gosec // build configuration is meant to be world-readable
			return fmt.Errorf("conform: write %s: %w", path, err)
		}
	}
	return nil
}

// compare reports how the repository's copy of name differs from original;
// deviates is false when the copy is identical.
func (r *Repository) compare(name string, original []byte) (d Deviation, deviates bool, err error) {
	body, err := r.read(name)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return Deviation{File: name, Kind: KindMissing}, true, nil
	case err != nil:
		return Deviation{}, false, err
	case bytes.Equal(body, original):
		return Deviation{}, false, nil
	}
	if v := standards.ConfigVersion(body); v != standards.ConfigVersion(original) {
		return Deviation{File: name, Kind: KindOutdated, Detail: fmt.Sprintf("has %q, standard is %s", v, standards.Version())}, true, nil
	}
	return Deviation{File: name, Kind: KindModified}, true, nil
}

func (r *Repository) read(name string) ([]byte, error) {
	body, err := os.ReadFile(filepath.Join(r.dir, name))
	if err != nil {
		return nil, fmt.Errorf("conform: read %s: %w", name, err)
	}
	return body, nil
}

func referencesStandard(agents []byte) bool {
	s := bufio.NewScanner(bytes.NewReader(agents))
	for i := 0; i < agentsHeadLines && s.Scan(); i++ {
		if strings.Contains(s.Text(), AgentsMarker) {
			return true
		}
	}
	return false
}
