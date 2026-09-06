# Architecture

## Why this exists

`systemd-analyze verify` is authoritative, but it needs the systemd binary
present at a version matching the units under test, attempts a D-Bus connection,
and follows references into units the caller doesn't own. Running it in a
container produces `Failed to create bus connection` and findings about
`dbus.socket`. See [systemd#15342](https://github.com/systemd/systemd/issues/15342).

`sdlint` reads unit files as text. Nothing else.

## Rule tiers

Rules are split by what input they require. The tier is expressed in the type
system so the loader cannot hand a rule context it did not ask for.

| Tier | Input | Example |
|------|-------|---------|
| T1 | one file's text | `Type=` is not a valid service type |
| T2 | a directory of units | `After=` names a unit that doesn't exist |
| T3 | a target rootfs or systemd version | `ExecStart=` binary is absent |

T1 is the bulk of the value and needs no configuration. T2 requires a search
path. T3 is always opt-in behind `--root` or `--target-version`, never required.

## Pipeline

```
paths → loader → []Unit ──┬─→ T1 rules ─┐
                          │             │
                    tree.Index ─────────┼─→ diag.Set → filter → reporter → exit
                          │             │
                    sysctx.Context ─────┘
```

Two invariants:

**Parse errors gate everything.** A file that fails to parse runs syntax rules
only. Emitting forty cascading semantic errors from one missing bracket is the
fastest way to lose a user.

**Rules never print.** They append to a `diag.Set`. Only reporters write output.
This is what makes JSON, SARIF, and GitHub annotations additive rather than a
retrofit.

## Why a custom lexer

`coreos/go-systemd`'s `unit` package deserializes into `{Section, Name, Value}`
triples and discards line and column information. A linter that cannot report
`file:line:col` is not a linter, so `internal/unitfile` implements its own
position-preserving lexer. go-systemd remains useful as a cross-check in tests.

`Directive` carries both `Value` and `Raw`. Rules needing semantics use `Value`;
rules needing an accurate caret column use `Raw` and `ValuePos`.

## Why the directive tables are generated

systemd's authoritative section/directive mapping lives in its source as a gperf
template. Hand-maintaining a copy guarantees drift: systemd-lsp's maintainer
names exactly this as the hard part of that project.

`internal/spec` is generated from that template by `go:generate` and committed,
so the build needs no network and contributors need no systemd checkout. The
source version is pinned in the generator and surfaced by `sdlint version`.
Supporting a new systemd release becomes a reviewable diff.

## Adding a rule

One file, one `init()` registration, one testdata directory:

```
internal/rules/svc/svc009.go
testdata/rules/SVC009/bad.service
testdata/rules/SVC009/bad.service.want
testdata/rules/SVC009/ok.service          # no .want means expect no findings
```

Rule metadata lives in the same file as its `Check` method. `sdlint explain`,
`sdlint rules --format=markdown`, and `docs/rules/*.md` all generate from it, so
documentation cannot drift from behaviour.

A meta-test asserts every registered rule ID has a testdata directory with at
least one positive and one negative case. Untested rules are a build failure.

## False positives are bugs

A linter people mute is worthless. Any rule that fires on a correct unit is a
defect, not a matter of taste. Rules whose false-positive rate exceeds roughly
one per twenty real-world units get demoted to `hint` or moved behind an opt-in
profile.
