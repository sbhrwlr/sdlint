# sdlint

Lint systemd unit files. No systemd, no root, no bus.

`systemd-analyze verify` needs the systemd binary installed at a matching
version, tries to open a D-Bus connection, and follows references into units you
don't own. That makes it awkward in CI, in containers, and useless for
cross-compiled images built on a machine that isn't the target.

`sdlint` is a single static binary that reads unit files as text.

> **Status: pre-alpha.** The CLI skeleton builds and runs. No rules are
> implemented yet. Not usable.

## Install

```
go install github.com/sbhrwlr/sdlint/cmd/sdlint@latest
```

## Usage

```
sdlint ./units/
sdlint nginx.service
sdlint                      # current directory
```

Exit codes: `0` clean, `1` findings at or above `--fail-on`, `2` usage error,
`3` internal error.

## Related tools

Use the right one for the job:

- **[systemd-analyze verify](https://www.freedesktop.org/software/systemd/man/latest/systemd-analyze.html)**
  — authoritative, since it is systemd. Use it when you can run it on the target
  system.
- **[systemdlint](https://github.com/priv-kweihmann/systemdlint)** — Python, the
  original offline linter, broad rule set.
- **[systemd-lsp](https://github.com/jfryy/systemd-lsp)** — editor integration
  with diagnostics, completion, and hover docs. Use it while writing units.

`sdlint` targets CI specifically: rule IDs, severity levels, per-rule disabling,
machine-readable output, and a binary with no runtime dependencies.

## License

MIT. See [LICENSE](LICENSE).
