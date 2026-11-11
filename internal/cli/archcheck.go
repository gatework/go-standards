package cli

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gatework/go-standards/internal/archcheck"
)

// Lister runs `go list` with the given arguments in the module root dir.
type Lister func(ctx context.Context, dir string, args ...string) ([]byte, error)

// GoList is the Lister that runs the go command.
func GoList(ctx context.Context, dir string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "go", append([]string{"list"}, args...)...) //nolint:gosec // runs the go tool with package patterns the operator named
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go list: %w", err)
	}
	return out, nil
}

// Archcheck enforces the module's declared package import boundaries (ARC-1)
// and returns the exit code: 1 lists one violation per line.
func Archcheck(ctx context.Context, args []string, list Lister, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("archcheck", flag.ContinueOnError)
	flags.SetOutput(stderr)
	dir := flags.String("dir", ".", "module root to check")
	rules := flags.String("rules", "architecture.yaml", "rule table, relative to -dir")
	if err := flags.Parse(args); err != nil {
		return ExitUsage
	}
	config, err := loadRules(filepath.Join(*dir, *rules))
	if err != nil {
		fmt.Fprintln(stderr, "archcheck:", err)
		return ExitUsage
	}
	patterns := flags.Args()
	if len(patterns) == 0 {
		patterns = []string{"./..."}
	}
	module, err := list(ctx, *dir, "-m")
	if err != nil {
		fmt.Fprintln(stderr, "archcheck:", err)
		return ExitUsage
	}
	out, err := list(ctx, *dir, append([]string{"-json=ImportPath,Imports,Deps"}, patterns...)...)
	if err != nil {
		fmt.Fprintln(stderr, "archcheck:", err)
		return ExitUsage
	}
	packages, err := archcheck.DecodePackages(bytes.NewReader(out))
	if err != nil {
		fmt.Fprintln(stderr, "archcheck:", err)
		return ExitUsage
	}
	violations := archcheck.Check(config, string(bytes.TrimSpace(module)), packages)
	if len(violations) > 0 {
		fmt.Fprintln(stderr, "architecture violations:")
		for _, v := range violations {
			fmt.Fprintln(stderr, " ", v)
		}
		return ExitDeviation
	}
	fmt.Fprintf(stdout, "PASS: %d packages satisfy %d rules\n", len(packages), len(config.Rules))
	return ExitOK
}

func loadRules(path string) (archcheck.Config, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return archcheck.Config{}, fmt.Errorf("read rules: %w", err)
	}
	config, err := archcheck.DecodeConfig(body)
	if err != nil {
		return archcheck.Config{}, fmt.Errorf("decode rules: %w", err)
	}
	return config, nil
}
