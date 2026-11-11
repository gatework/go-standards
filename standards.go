package standards

import (
	"bytes"
	_ "embed"
	"regexp"
)

var (
	//go:embed STANDARD.md
	standard []byte
	//go:embed golangci.yml
	lintConfig []byte
	//go:embed standard.mk
	makeTargets []byte
)

var versionLine = regexp.MustCompile(`(?m)^# standard-version: (\S+)$`)

// Text returns a copy of the normative standard.
func Text() []byte { return bytes.Clone(standard) }

// LintConfig returns a copy of the shared golangci-lint configuration.
func LintConfig() []byte { return bytes.Clone(lintConfig) }

// MakeTargets returns a copy of the shared make targets (standard.mk).
func MakeTargets() []byte { return bytes.Clone(makeTargets) }

// Version is the standard version stamped in the lint configuration.
func Version() string { return ConfigVersion(lintConfig) }

// ConfigVersion returns the standard version stamped in a distributed
// file, or "" when it carries none.
func ConfigVersion(config []byte) string {
	if m := versionLine.FindSubmatch(config); m != nil {
		return string(m[1])
	}
	return ""
}
