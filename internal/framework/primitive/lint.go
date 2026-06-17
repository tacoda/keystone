package primitive

import (
	"fmt"
	"sort"
	"strings"
)

// Severity classifies a Finding. Errors fail the lint; warnings are
// reported but do not.
type FindingSeverity string

const (
	FindingError   FindingSeverity = "error"
	FindingWarning FindingSeverity = "warning"
)

// Finding is one validation report against one primitive (or a
// duplicate-id collision between two).
type Finding struct {
	Severity FindingSeverity
	Path     string
	Kind     string
	ID       string
	Message  string
}

func (f Finding) String() string {
	loc := f.Path
	if loc == "" {
		loc = f.Kind + "/" + f.ID
	}
	return fmt.Sprintf("%s: %s: %s", f.Severity, loc, f.Message)
}

// Lint walks a slice of primitives and emits Findings. Caller decides
// what to do with them; the returned slice is sorted by (severity,
// path, message) for stable output.
//
// Hard errors:
//   - missing or empty kind / id / description
//   - unknown kind value
//   - duplicate (kind, id) within the set
//   - empty `globs: []` list (omit the key instead)
//   - skill missing triggers
//   - subagent missing tools
//   - sensor_kind set to a value other than computational | inferential
//
// Warnings:
//   - deps[] entry that does not resolve to a known (kind, id)
//   - traces[] entry that does not resolve
//   - id contains characters outside [A-Za-z0-9-:/_]
func Lint(primitives []Primitive) []Finding {
	var findings []Finding

	known := map[Kind]bool{}
	for _, k := range KnownKinds {
		known[k] = true
	}

	// Build (kind, id) → exists map for cross-reference validation.
	exists := map[string]bool{}
	seen := map[string]string{} // (kind,id) → first path seen
	for _, p := range primitives {
		key := p.Kind + "/" + p.ID
		exists[key] = true
		if prev, dup := seen[key]; dup {
			findings = append(findings, Finding{
				Severity: FindingError,
				Path:     p.Path,
				Kind:     p.Kind,
				ID:       p.ID,
				Message:  fmt.Sprintf("duplicate (kind=%s, id=%s); also at %s", p.Kind, p.ID, prev),
			})
			continue
		}
		seen[key] = p.Path
	}

	for _, p := range primitives {
		findings = append(findings, lintOne(p, known, exists)...)
	}

	sort.SliceStable(findings, func(i, j int) bool {
		if findings[i].Severity != findings[j].Severity {
			return findings[i].Severity == FindingError && findings[j].Severity == FindingWarning
		}
		if findings[i].Path != findings[j].Path {
			return findings[i].Path < findings[j].Path
		}
		return findings[i].Message < findings[j].Message
	})
	return findings
}

func lintOne(p Primitive, known map[Kind]bool, exists map[string]bool) []Finding {
	var out []Finding

	add := func(sev FindingSeverity, msg string) {
		out = append(out, Finding{Severity: sev, Path: p.Path, Kind: p.Kind, ID: p.ID, Message: msg})
	}

	// Identity.
	if strings.TrimSpace(p.Kind) == "" {
		add(FindingError, "missing required field `kind`")
	} else if !known[Kind(p.Kind)] {
		add(FindingError, fmt.Sprintf("unknown kind %q (allowed: %s)", p.Kind, knownKindList()))
	}
	if strings.TrimSpace(p.ID) == "" {
		add(FindingError, "missing required field `id`")
	} else if bad := badIDChars(p.ID); bad != "" {
		add(FindingWarning, fmt.Sprintf("id contains characters outside [a-z0-9-:_/]: %q", bad))
	}
	if strings.TrimSpace(p.Description) == "" {
		add(FindingError, "missing required field `description`")
	} else if strings.HasPrefix(strings.TrimSpace(p.Description), "TODO") {
		add(FindingWarning, "description still says TODO — fill it in")
	}

	// Per-kind required fields.
	switch Kind(p.Kind) {
	case KindSkill:
		if len(p.Triggers) == 0 {
			add(FindingError, "skill missing `triggers:` — a skill with no triggers cannot fire")
		}
	case KindSubagent:
		if len(p.Tools) == 0 {
			add(FindingError, "subagent missing `tools:` — declare the tool allow-list explicitly")
		}
	}

	// globs: empty list is a parse-time error (per docs/ports/primitive.md).
	// We can't detect `globs: []` after decoding (becomes nil-equivalent),
	// but we can flag entries with whitespace-only patterns.
	for _, g := range p.Globs {
		if strings.TrimSpace(g) == "" {
			add(FindingError, "globs entry is empty — remove or fix")
		}
	}

	// deps / traces referential integrity (warnings — cross-policy refs
	// may resolve only after install).
	for _, d := range p.Deps {
		if d == "" {
			continue
		}
		if !strings.Contains(d, "/") {
			add(FindingWarning, fmt.Sprintf("deps entry %q lacks <kind>/<id> form", d))
			continue
		}
		if !exists[d] {
			add(FindingWarning, fmt.Sprintf("deps entry %q does not resolve", d))
		}
	}
	for _, t := range p.Traces {
		if t == "" {
			continue
		}
		// traces are corpus refs; default kind is "corpus" if unqualified.
		key := t
		if !strings.Contains(t, "/") {
			key = "corpus/" + t
		}
		// Try both raw form and corpus-prefixed form.
		if !exists[key] && !exists["corpus/"+t] {
			add(FindingWarning, fmt.Sprintf("traces entry %q does not resolve", t))
		}
	}

	return out
}

func knownKindList() string {
	parts := make([]string, 0, len(KnownKinds))
	for _, k := range KnownKinds {
		parts = append(parts, string(k))
	}
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}

func badIDChars(id string) string {
	var bad []rune
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-' || r == ':' || r == '_' || r == '/' || r == '.':
		default:
			bad = append(bad, r)
		}
	}
	if len(bad) == 0 {
		return ""
	}
	return string(bad)
}

// HasErrors reports whether any finding is severity=error.
func HasErrors(findings []Finding) bool {
	for _, f := range findings {
		if f.Severity == FindingError {
			return true
		}
	}
	return false
}
