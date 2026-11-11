// Command archcheck enforces a module's declared package import boundaries
// (STANDARD.md ARC-1).
//
// From the module root:
//
//	go run github.com/gatework/go-standards/cmd/archcheck@<version> [-dir .] [-rules architecture.yaml] [packages]
//
// It reads the production import graph from `go list -json` (default ./...)
// and exits 1 with one line per violation.
package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/gatework/go-standards/internal/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	code := cli.Archcheck(ctx, os.Args[1:], cli.GoList, os.Stdout, os.Stderr)
	stop()
	os.Exit(code)
}
