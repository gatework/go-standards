package standards_test

import (
	"bytes"
	"regexp"
	"testing"

	standards "github.com/gatework/go-standards"
)

func TestVersionMatchesStandard(t *testing.T) {
	t.Parallel()
	m := regexp.MustCompile(`版本：(\d+\.\d+\.\d+)`).FindSubmatch(standards.Text())
	if m == nil {
		t.Fatal("STANDARD.md has no version line")
	}
	if got := standards.Version(); got != string(m[1]) {
		t.Fatalf("golangci.yml stamps %q, STANDARD.md is %s", got, m[1])
	}
	if got := standards.ConfigVersion(standards.MakeTargets()); got != string(m[1]) {
		t.Fatalf("standard.mk stamps %q, STANDARD.md is %s", got, m[1])
	}
}

func TestCopiesAreDetached(t *testing.T) {
	t.Parallel()
	a := standards.LintConfig()
	a[0] = 'X'
	if bytes.Equal(a, standards.LintConfig()) {
		t.Fatal("LintConfig shares its backing array")
	}
}

func TestConfigVersion(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		config, want string
	}{
		"stamped":   {"version: \"2\"\n# standard-version: 1.2.3\n", "1.2.3"},
		"unstamped": {"version: \"2\"\n", ""},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := standards.ConfigVersion([]byte(tc.config)); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}
