// Package diag defines the diagnostic types that rules emit and reporters render.
//
// Rules never write to stdout. They append to a Set. Only reporters produce
// output. This separation is what makes alternate output formats (JSON, SARIF,
// GitHub annotations) a matter of adding a reporter rather than touching rules.
package diag

import (
	"fmt"
	"sort"
)

// Severity orders from most to least serious. The zero value is deliberately
// invalid so an unset severity is caught rather than silently treated as style.
type Severity int

const (
	Invalid Severity = iota
	// Error means systemd will reject the unit or it provably cannot work.
	Error
	// Warning means the unit is legal but this is near-certainly a bug.
	Warning
	// Hint means a real footgun that nonetheless has legitimate uses.
	Hint
	// Style is opinion: hardening, documentation, conventions.
	Style
)

func (s Severity) String() string {
	switch s {
	case Error:
		return "error"
	case Warning:
		return "warning"
	case Hint:
		return "hint"
	case Style:
		return "style"
	default:
		return "invalid"
	}
}

// ParseSeverity maps a --fail-on value to a Severity.
func ParseSeverity(s string) (Severity, error) {
	switch s {
	case "error":
		return Error, nil
	case "warning":
		return Warning, nil
	case "hint":
		return Hint, nil
	case "style":
		return Style, nil
	default:
		return Invalid, fmt.Errorf("unknown severity %q", s)
	}
}

// Position is a 1-based location in a unit file. Col may be 0 when a finding
// applies to a whole line rather than a span within it.
type Position struct {
	File string
	Line int
	Col  int
}

func (p Position) String() string {
	if p.Col > 0 {
		return fmt.Sprintf("%s:%d:%d", p.File, p.Line, p.Col)
	}
	return fmt.Sprintf("%s:%d", p.File, p.Line)
}

// Diagnostic is a single finding. RuleID ties it back to the rule that produced
// it so `sdlint explain <RuleID>` can expand on it.
type Diagnostic struct {
	RuleID   string
	Severity Severity
	Message  string
	Pos      Position
	// Source is the offending line, used to render a caret. Optional.
	Source string
}

// Set collects diagnostics across every file in a run.
type Set struct {
	items []Diagnostic
}

func (s *Set) Add(d Diagnostic) { s.items = append(s.items, d) }

func (s *Set) Items() []Diagnostic { return s.items }

func (s *Set) Len() int { return len(s.items) }

// Sort orders diagnostics by file, then line, then column, then rule ID so
// output is stable across runs. Golden-file tests depend on this.
func (s *Set) Sort() {
	sort.SliceStable(s.items, func(i, j int) bool {
		a, b := s.items[i], s.items[j]
		if a.Pos.File != b.Pos.File {
			return a.Pos.File < b.Pos.File
		}
		if a.Pos.Line != b.Pos.Line {
			return a.Pos.Line < b.Pos.Line
		}
		if a.Pos.Col != b.Pos.Col {
			return a.Pos.Col < b.Pos.Col
		}
		return a.RuleID < b.RuleID
	})
}

// Counts returns how many diagnostics exist at each severity.
func (s *Set) Counts() map[Severity]int {
	m := make(map[Severity]int)
	for _, d := range s.items {
		m[d.Severity]++
	}
	return m
}

// ExceedsThreshold reports whether any diagnostic is at or above the given
// severity. Because Error is the lowest numeric value, "at or above" is <=.
func (s *Set) ExceedsThreshold(threshold Severity) bool {
	for _, d := range s.items {
		if d.Severity <= threshold {
			return true
		}
	}
	return false
}
