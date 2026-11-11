// Command stdsync distributes and verifies the shared gatework Go standard
// files in a repository.
package main

import (
	"os"

	"github.com/gatework/go-standards/internal/cli"
)

func main() {
	os.Exit(cli.Stdsync(os.Args[1:], os.Stdout, os.Stderr))
}
