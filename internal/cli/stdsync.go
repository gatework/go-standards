package cli

import (
	"fmt"
	"io"

	standards "github.com/gatework/go-standards"
	"github.com/gatework/go-standards/internal/conform"
)

// Exit codes.
const (
	ExitOK        = 0
	ExitDeviation = 1
	ExitUsage     = 2
)

const usage = `usage: stdsync <command> [repository...]

commands:
  check    report where repositories deviate from the standard (exit 1 if any)
  write    replace each repository's .golangci.yml and standard.mk with the shared ones
  version  print the standard version
  config   print the shared golangci-lint configuration

Repositories default to the current directory.
`

// Stdsync executes one stdsync invocation and returns its exit code.
func Stdsync(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return ExitUsage
	}
	dirs := args[1:]
	if len(dirs) == 0 {
		dirs = []string{"."}
	}
	switch args[0] {
	case "check":
		return check(dirs, stdout, stderr)
	case "write":
		return write(dirs, stdout, stderr)
	case "version":
		fmt.Fprintln(stdout, standards.Version())
		return ExitOK
	case "config":
		_, _ = stdout.Write(standards.LintConfig())
		return ExitOK
	case "help", "-h", "--help":
		fmt.Fprint(stdout, usage)
		return ExitOK
	}
	fmt.Fprintf(stderr, "stdsync: unknown command %q\n%s", args[0], usage)
	return ExitUsage
}

func check(dirs []string, stdout, stderr io.Writer) int {
	code := ExitOK
	for _, dir := range dirs {
		r, err := conform.Open(dir)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return ExitUsage
		}
		deviations, err := r.Check()
		if err != nil {
			fmt.Fprintln(stderr, err)
			return ExitUsage
		}
		for _, d := range deviations {
			fmt.Fprintf(stdout, "%s: %s\n", r.Dir(), d)
			code = ExitDeviation
		}
	}
	return code
}

func write(dirs []string, stdout, stderr io.Writer) int {
	for _, dir := range dirs {
		r, err := conform.Open(dir)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return ExitUsage
		}
		if err := r.Write(); err != nil {
			fmt.Fprintln(stderr, err)
			return ExitUsage
		}
		fmt.Fprintf(stdout, "%s: wrote %s, %s (standard %s)\n", r.Dir(), conform.LintFile, conform.MakeFile, standards.Version())
	}
	return ExitOK
}
