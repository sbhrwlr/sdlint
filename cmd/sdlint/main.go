// Command sdlint lints systemd unit files without a running systemd.
package main

import (
	"os"

	"github.com/sbhrwlr/sdlint/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
