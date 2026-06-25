package primitive

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// tagPattern enforces kebab-case for tag values — lowercase letters,
// digits, and hyphens; must start with a letter; max 64 chars.
var tagPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)

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
//   - persona missing tools (persona wraps subagent — same constraint)
//   - projection-target collision: a framework wrapper and its agent
//     escape hatch sharing the same id project to the same .claude/
//     path (e.g. persona `foo` and subagent `foo` both → .claude/agents/foo.md)
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

	findings = append(findings, lintProjectionCollisions(primitives)...)

	// Compose-time violations (unknown concern id, leaf violation,
	// duplicate include) are lint errors — they break the indexer's
	// ability to produce a deterministic merged descriptor.
	_, composeErrs := Compose(primitives)
	for _, e := range composeErrs {
		findings = append(findings, Finding{
			Severity: FindingError,
			Path:     e.Path,
			ID:       e.ID,
			Message:  e.Message,
		})
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
	lintKindAndMode(p, known, add)
	lintIDAndDescription(p, add)
	lintKindFields(p, add)
	lintTagsAndGlobs(p, add)
	out = append(out, lintRefs(p, exists)...)
	return out
}

// lintKindAndMode validates the `kind:` value (closed set; `rule` is rejected
// with a guide-pointing hint) and the cross-cutting `mode:` value.
func lintKindAndMode(p Primitive, known map[Kind]bool, add func(FindingSeverity, string)) {
	if strings.TrimSpace(p.Kind) == "" {
		add(FindingError, "missing required field `kind`")
	} else if p.Kind == "rule" {
		add(FindingError, "rule is not a kind — author a guide (rule is a projection-target name)")
	} else if !known[Kind(p.Kind)] {
		add(FindingError, fmt.Sprintf("unknown kind %q (allowed: %s)", p.Kind, knownKindList()))
	}
	switch p.Mode {
	case "", "computational", "inferential":
	default:
		add(FindingError, fmt.Sprintf("mode %q invalid (computational | inferential)", p.Mode))
	}
}

// lintIDAndDescription validates the two remaining identity fields.
func lintIDAndDescription(p Primitive, add func(FindingSeverity, string)) {
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
}

// lintKindFields validates the per-kind required fields and contracts.
func lintKindFields(p Primitive, add func(FindingSeverity, string)) {
	switch Kind(p.Kind) {
	case KindGuide:
		lintGuideTier(p, add)
	case KindSkill:
		if len(p.Triggers) == 0 {
			add(FindingError, "skill missing `triggers:` — a skill with no triggers cannot fire")
		}
	case KindAgent:
		if len(p.Tools) == 0 {
			add(FindingError, "agent missing `tools:` — declare the tool allow-list explicitly")
		}
	case KindSensor, KindHook:
		lintActionContract(p, add)
	case KindTool:
		lintToolContract(p, add)
	}
}

// lintGuideTier validates an inferential guide's `tier:` — iron-law |
// golden-rule | preference (the default; empty = preference).
func lintGuideTier(p Primitive, add func(FindingSeverity, string)) {
	switch p.Tier {
	case "", string(TierIronLaw), string(TierGoldenRule), string(TierPreference):
	default:
		add(FindingError, fmt.Sprintf("guide tier %q invalid (iron-law | golden-rule | preference)", p.Tier))
	}
}

// lintToolContract enforces a tool's handler + transport. A tool is an
// author-defined callable: it needs a `run:` handler and a `transport:` of
// cli | mcp | plugin (empty defaults to cli).
func lintToolContract(p Primitive, add func(FindingSeverity, string)) {
	switch p.Transport {
	case "", "cli", "mcp", "plugin":
	default:
		add(FindingError, fmt.Sprintf("tool transport %q invalid (cli | mcp | plugin)", p.Transport))
	}
	if strings.TrimSpace(p.Run) == "" {
		add(FindingError, "tool missing `run:` — the handler the agent invokes")
	}
}

// lintTagsAndGlobs enforces kebab-case tags and rejects whitespace-only globs.
// Tags are an orthogonal taxonomy surfaced by `keystone list --tag X`; keeping
// the shape predictable keeps them grep-friendly. A `globs: []` empty list is
// a parse-time error; here we flag entries that are whitespace-only.
func lintTagsAndGlobs(p Primitive, add func(FindingSeverity, string)) {
	for _, t := range p.Tags {
		if !tagPattern.MatchString(t) {
			add(FindingError, fmt.Sprintf("tag %q is not kebab-case (lowercase letters, digits, hyphens; must start with a letter)", t))
		}
	}
	for _, g := range p.Globs {
		if strings.TrimSpace(g) == "" {
			add(FindingError, "globs entry is empty — remove or fix")
		}
	}
}

// lintActionContract enforces the computational/inferential action shape on
// a sensor or hook. Computational → a `run:` shell command/script.
// Inferential → an `agent:` to dispatch plus a `returns:` structured-result
// schema. The two are mutually exclusive. Default mode for both kinds is
// computational. `add` is the caller's finding accumulator.
func lintActionContract(p Primitive, add func(FindingSeverity, string)) {
	if strings.TrimSpace(p.Run) != "" && strings.TrimSpace(p.Agent) != "" {
		add(FindingError, "`run:` and `agent:` are mutually exclusive — a sensor/hook is computational (run) or inferential (agent), not both")
	}
	if p.Mode == "inferential" {
		if strings.TrimSpace(p.Agent) == "" {
			add(FindingError, "inferential sensor/hook missing `agent:` — name the agent to dispatch")
		}
		if strings.TrimSpace(p.Returns) == "" {
			add(FindingError, "inferential sensor/hook missing `returns:` — declare the structured-result schema")
		}
		return
	}
	if strings.TrimSpace(p.Run) == "" {
		add(FindingError, "computational sensor/hook missing `run:` — the shell command/script to execute")
	}
}

// lintRefs validates a primitive's association fields — the frontmatter
// links that wire concepts together. Dangling targets are warnings:
// cross-policy refs may resolve only after install. Split out of lintOne
// to keep each function's complexity in check.
func lintRefs(p Primitive, exists map[string]bool) []Finding {
	// Each association field resolves a ref against a set of candidate
	// key-prefixes ("" = the raw ref). corpus defaults to the corpus/
	// kind; produces/consumes target documents; supersedes targets a
	// same-kind id. Data-driven so the function stays flat.
	checks := []struct {
		field    string
		refs     []string
		prefixes []string
	}{
		{"corpus", p.Corpus, []string{"", "corpus/"}},
		{"produces", p.Produces, []string{"document/", ""}},
		{"consumes", p.Consumes, []string{"document/", ""}},
		{"supersedes", p.Supersedes, []string{p.Kind + "/", ""}},
	}
	var out []Finding
	for _, c := range checks {
		for _, ref := range c.refs {
			if ref == "" || refResolves(exists, ref, c.prefixes) {
				continue
			}
			out = append(out, Finding{
				Severity: FindingWarning, Path: p.Path, Kind: p.Kind, ID: p.ID,
				Message: fmt.Sprintf("%s entry %q does not resolve", c.field, ref),
			})
		}
	}
	return out
}

// refResolves reports whether ref exists under any of the candidate
// prefixes (prefix+ref) in the (kind/id) set.
func refResolves(exists map[string]bool, ref string, prefixes []string) bool {
	for _, pre := range prefixes {
		if exists[pre+ref] {
			return true
		}
	}
	return false
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

// lintProjectionCollisions reports two or more primitives that would
// project to the same .claude/ target path. Framework wrappers
// (persona/action/playbook) intentionally share projection targets
// with their agent counterparts (subagent/command/skill); the author
// must pick one authoring layer per id.
func lintProjectionCollisions(primitives []Primitive) []Finding {
	type ref struct {
		kind string
		id   string
		path string
	}
	byTarget := map[string][]ref{}
	for _, p := range primitives {
		target := ProjectionRelPath(p)
		if target == "" {
			continue
		}
		byTarget[target] = append(byTarget[target], ref{kind: p.Kind, id: p.ID, path: p.Path})
	}

	var out []Finding
	targets := make([]string, 0, len(byTarget))
	for t := range byTarget {
		targets = append(targets, t)
	}
	sort.Strings(targets)
	for _, t := range targets {
		refs := byTarget[t]
		if len(refs) < 2 {
			continue
		}
		// Build a "kind=id at path" list for the message; emit one
		// finding per colliding primitive so each path surfaces.
		summary := make([]string, 0, len(refs))
		for _, r := range refs {
			summary = append(summary, fmt.Sprintf("%s=%s", r.kind, r.id))
		}
		sort.Strings(summary)
		msg := fmt.Sprintf("projection collision at %s — %s map to the same target; rename one (framework wrappers share targets with their agent escape hatches by design)", t, strings.Join(summary, " + "))
		for _, r := range refs {
			out = append(out, Finding{
				Severity: FindingError,
				Path:     r.path,
				Kind:     r.kind,
				ID:       r.id,
				Message:  msg,
			})
		}
	}
	return out
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
