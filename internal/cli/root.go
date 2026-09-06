// Package cli wires flags to the linter and maps results to exit codes.
package cli

import (
	"flag"
	"fmt"
	"io"

	"github.com/sbhrwlr/sdlint/internal/diag"
)

// Exit codes. Keeping usage errors (2) distinct from internal errors (3) lets
// CI tell "your units are broken" apart from "the linter is broken".
const (
	ExitOK       = 0
	ExitFindings = 1
	ExitUsage    = 2
	ExitInternal = 3
)

const tagline = "Lint systemd unit files. No systemd, no root, no bus."

// Options holds resolved configuration for one run.
type Options struct {
	Paths  []string
	Format string
	FailOn diag.Severity
}

// Run parses arguments and executes the linter. It returns a process exit code
// rather than calling os.Exit so it stays testable.
func Run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("sdlint", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		format      = fs.String("format", "text", "output format: text, json, github, sarif")
		failOn      = fs.String("fail-on", "warning", "minimum severity that fails the run: error, warning, hint, style")
		showVersion = fs.Bool("version", false, "print version information and exit")
	)

	fs.Usage = func() {
		fmt.Fprintf(stderr, "%s\n\nusage: sdlint [flags] [paths...]\n\n", tagline)
		fmt.Fprintf(stderr, "With no paths, the current directory is linted.\n\nflags:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return ExitUsage
	}

	if *showVersion {
		fmt.Fprintln(stdout, VersionString())
		return ExitOK
	}

	threshold, err := diag.ParseSeverity(*failOn)
	if err != nil {
		fmt.Fprintf(stderr, "sdlint: %v\n", err)
		return ExitUsage
	}

	paths := fs.Args()
	if len(paths) == 0 {
		paths = []string{"."}
	}

	opts := Options{Paths: paths, Format: *format, FailOn: threshold}
	return lint(opts, stdout, stderr)
}

// lint is the pipeline entry point. Loading, rule execution, and reporting land
// here as they are implemented.
func lint(opts Options, stdout, stderr io.Writer) int {
	var set diag.Set

	// TODO(lexer): parse each path in opts.Paths into a *unitfile.Unit.
	// TODO(rules): run the registered rule set against each unit.
	// TODO(report): render set according to opts.Format.

	set.Sort()

	if set.Len() == 0 {
		fmt.Fprintln(stdout, "no findings")
		return ExitOK
	}
	if set.ExceedsThreshold(opts.FailOn) {
		return ExitFindings
	}
	return ExitOK
}
