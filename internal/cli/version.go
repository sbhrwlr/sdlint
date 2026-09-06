package cli

import "fmt"

// Set at build time via -ldflags. See the Makefile.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// systemdVersion records which systemd release the generated directive tables
// were derived from. Reporting it matters: a user seeing an unexpected
// "unknown directive" needs to know which systemd this binary knows about.
const systemdVersion = "unset"

// VersionString renders the full version block.
func VersionString() string {
	return fmt.Sprintf("sdlint %s\ncommit:  %s\nbuilt:   %s\nsystemd: %s",
		version, commit, date, systemdVersion)
}
