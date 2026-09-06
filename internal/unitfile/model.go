// Package unitfile parses systemd unit files while preserving source positions.
//
// coreos/go-systemd's unit package deserializes into {Section, Name, Value}
// triples and discards line and column information. A linter cannot report
// file:line:col without that, so this package implements its own lexer rather
// than wrapping go-systemd.
package unitfile

// Type is the kind of unit, derived from the filename suffix.
type Type string

const (
	TypeService   Type = "service"
	TypeSocket    Type = "socket"
	TypeTimer     Type = "timer"
	TypeTarget    Type = "target"
	TypeMount     Type = "mount"
	TypeAutomount Type = "automount"
	TypeSwap      Type = "swap"
	TypePath      Type = "path"
	TypeSlice     Type = "slice"
	TypeScope     Type = "scope"
	TypeDevice    Type = "device"
	TypeUnknown   Type = ""
)

// KnownTypes lists every valid unit suffix. Used by NAM001.
var KnownTypes = []Type{
	TypeService, TypeSocket, TypeTimer, TypeTarget, TypeMount,
	TypeAutomount, TypeSwap, TypePath, TypeSlice, TypeScope, TypeDevice,
}

// Position is a 1-based source location within a unit file.
type Position struct {
	Line int
	Col  int
}

// Directive is a single key=value assignment.
//
// Raw holds the value exactly as written, before unquoting or continuation
// joining. Rules that need accurate column offsets for a caret use Raw and
// ValuePos; rules that need semantics use Value.
type Directive struct {
	Section  string
	Name     string
	Value    string
	Raw      string
	Pos      Position
	ValuePos Position
}

// Section is a [Header] and the directives beneath it.
type Section struct {
	Name       string
	Pos        Position
	Directives []Directive
}

// Unit is one parsed unit file.
//
// When ParseError is non-nil the file could not be fully parsed. Callers must
// run only syntax rules in that case: reporting dozens of cascading semantic
// errors from a single missing bracket is the fastest way to lose a user.
type Unit struct {
	Path       string
	Name       string
	Type       Type
	Instance   string
	Sections   []Section
	IsDropIn   bool
	ParseError error
}

// Get returns every directive with the given name in the given section, in
// source order. Returns nil when absent.
func (u *Unit) Get(section, name string) []Directive {
	var out []Directive
	for _, s := range u.Sections {
		if !equalFold(s.Name, section) {
			continue
		}
		for _, d := range s.Directives {
			if equalFold(d.Name, name) {
				out = append(out, d)
			}
		}
	}
	return out
}

// First returns the first directive with the given name, and whether it exists.
func (u *Unit) First(section, name string) (Directive, bool) {
	got := u.Get(section, name)
	if len(got) == 0 {
		return Directive{}, false
	}
	return got[0], true
}

// Has reports whether the directive is present at least once.
func (u *Unit) Has(section, name string) bool {
	return len(u.Get(section, name)) > 0
}

// HasSection reports whether the named section is present.
func (u *Unit) HasSection(name string) bool {
	for _, s := range u.Sections {
		if equalFold(s.Name, name) {
			return true
		}
	}
	return false
}

// IsTemplate reports whether this is a template unit (foo@.service).
func (u *Unit) IsTemplate() bool {
	n := u.Name
	for i := 0; i < len(n); i++ {
		if n[i] == '@' {
			return i+1 < len(n) && n[i+1] == '.'
		}
	}
	return false
}

// equalFold compares ASCII strings case-insensitively. systemd section and
// directive names are ASCII, so this avoids pulling in unicode folding.
func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if 'A' <= ca && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if 'A' <= cb && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}
