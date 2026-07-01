package primitive

import (
	"strings"
	"testing"
)

func find(t *testing.T, fs []Finding, severity FindingSeverity, msgContains string) bool {
	t.Helper()
	for _, f := range fs {
		if f.Severity == severity && strings.Contains(f.Message, msgContains) {
			return true
		}
	}
	return false
}

func TestLint_RequiredFields(t *testing.T) {
	ps := []Primitive{
		{Frontmatter: Frontmatter{Kind: "guide", ID: "p/x"}},        // missing description
		{Frontmatter: Frontmatter{Kind: "skill", Description: "x"}}, // missing id
		{Frontmatter: Frontmatter{ID: "x", Description: "x"}},       // missing kind
	}
	fs := Lint(ps)
	if !find(t, fs, FindingError, "missing required field `description`") {
		t.Error("expected missing-description error")
	}
	if !find(t, fs, FindingError, "missing required field `id`") {
		t.Error("expected missing-id error")
	}
	if !find(t, fs, FindingError, "missing required field `kind`") {
		t.Error("expected missing-kind error")
	}
}

func TestLint_UnknownKind(t *testing.T) {
	ps := []Primitive{
		{Frontmatter: Frontmatter{Kind: "bogus", ID: "x", Description: "d"}},
	}
	fs := Lint(ps)
	if !find(t, fs, FindingError, "unknown kind") {
		t.Errorf("expected unknown-kind error, got %v", fs)
	}
}

// TestLint_RetiredKindIsUnknown — the collapse/2.x kinds that 3.0 dropped
// (action→command, persona/subagent→agent) are simply unknown.
func TestLint_RetiredKindIsUnknown(t *testing.T) {
	for _, old := range []string{"action", "persona", "subagent"} {
		t.Run(old, func(t *testing.T) {
			fs := Lint([]Primitive{{Frontmatter: Frontmatter{Kind: old, ID: "x", Description: "d"}}})
			if !find(t, fs, FindingError, "unknown kind") {
				t.Errorf("kind %q: expected unknown-kind error, got %v", old, fs)
			}
		})
	}
}

// TestLint_RuleIsNotAKind — `rule` is a projection-target name, not an
// authorable kind. Lint rejects it with a guide-pointing hint.
func TestLint_RuleIsNotAKind(t *testing.T) {
	fs := Lint([]Primitive{{Frontmatter: Frontmatter{Kind: "rule", ID: "x", Description: "d"}}})
	if !find(t, fs, FindingError, "rule is not a kind") {
		t.Errorf("expected rule-is-not-a-kind error, got %v", fs)
	}
}

// TestLint_GuideTier — an inferential guide's `tier:` must be iron-law,
// golden-rule, or preference (empty = preference default).
func TestLint_GuideTier(t *testing.T) {
	for _, tier := range []string{"", "iron-law", "golden-rule", "preference"} {
		fs := Lint([]Primitive{{Frontmatter: Frontmatter{Kind: "guide", ID: "p/x", Description: "d", Tier: tier}}})
		if find(t, fs, FindingError, "tier") {
			t.Errorf("tier %q: expected no tier error, got %v", tier, fs)
		}
	}
	bad := Lint([]Primitive{{Frontmatter: Frontmatter{Kind: "guide", ID: "p/y", Description: "d", Tier: "critical"}}})
	if !find(t, bad, FindingError, "tier") {
		t.Errorf("expected invalid-tier error, got %v", bad)
	}
}

// TestLint_ToolContract — a tool needs a `run:` handler and a valid
// `transport:` (cli | mcp | plugin | "").
func TestLint_ToolContract(t *testing.T) {
	missingRun := Lint([]Primitive{{Frontmatter: Frontmatter{Kind: "tool", ID: "grep", Description: "d", Transport: "cli"}}})
	if !find(t, missingRun, FindingError, "run") {
		t.Errorf("expected tool-missing-run error, got %v", missingRun)
	}
	badTransport := Lint([]Primitive{{Frontmatter: Frontmatter{Kind: "tool", ID: "grep", Description: "d", Run: "./x.sh", Transport: "carrier-pigeon"}}})
	if !find(t, badTransport, FindingError, "transport") {
		t.Errorf("expected invalid-transport error, got %v", badTransport)
	}
	clean := Lint([]Primitive{{Frontmatter: Frontmatter{Kind: "tool", ID: "grep", Description: "d", Run: "./x.sh", Transport: "mcp"}}})
	if HasErrors(clean) {
		t.Errorf("expected clean tool, got %v", clean)
	}
}

// TestLint_ModeValue — `mode:` must be computational, inferential, or empty.
func TestLint_ModeValue(t *testing.T) {
	fs := Lint([]Primitive{{Frontmatter: Frontmatter{Kind: "guide", ID: "p/x", Description: "d", Mode: "bogus"}}})
	if !find(t, fs, FindingError, "mode") {
		t.Errorf("expected invalid-mode error, got %v", fs)
	}
	clean := Lint([]Primitive{{Frontmatter: Frontmatter{Kind: "guide", ID: "p/y", Description: "d", Mode: "inferential"}}})
	if find(t, clean, FindingError, "mode") {
		t.Errorf("did not expect a mode error for a valid mode, got %v", clean)
	}
}

// TestLint_SensorContract — a sensor is mode-driven: computational fires a
// `run:` check; inferential is a review needing a `returns:` verdict schema.
func TestLint_SensorContract(t *testing.T) {
	infMissing := Lint([]Primitive{{Frontmatter: Frontmatter{
		Kind: "sensor", ID: "review", Description: "d", Mode: "inferential"}}})
	if !find(t, infMissing, FindingError, "returns") {
		t.Errorf("expected inferential sensor to require returns, got %v", infMissing)
	}
	compMissing := Lint([]Primitive{{Frontmatter: Frontmatter{
		Kind: "sensor", ID: "build", Description: "d", Mode: "computational"}}})
	if !find(t, compMissing, FindingError, "run") {
		t.Errorf("expected computational sensor to require run, got %v", compMissing)
	}
	cleanInf := Lint([]Primitive{{Frontmatter: Frontmatter{
		Kind: "sensor", ID: "ok", Description: "d", Mode: "inferential", Returns: "findings"}}})
	if HasErrors(cleanInf) {
		t.Errorf("expected clean inferential sensor, got %v", cleanInf)
	}
	cleanComp := Lint([]Primitive{{Frontmatter: Frontmatter{
		Kind: "sensor", ID: "build2", Description: "d", Mode: "computational", Run: "go test ./...", Event: "pre-verify"}}})
	if HasErrors(cleanComp) {
		t.Errorf("expected clean computational sensor, got %v", cleanComp)
	}
}

// TestLint_SensorActionContract — a sensor's mode contract: computational
// requires `run:`; inferential requires a `returns:` verdict schema (the
// sensor body is the review prompt — no separate agent: needed).
func TestLint_SensorActionContract(t *testing.T) {
	infMissing := Lint([]Primitive{{Frontmatter: Frontmatter{
		Kind: "sensor", ID: "on-gate-review", Description: "d", Mode: "inferential", Event: "on-gate"}}})
	if !find(t, infMissing, FindingError, "returns") {
		t.Errorf("expected inferential sensor to require returns, got %v", infMissing)
	}
	compMissing := Lint([]Primitive{{Frontmatter: Frontmatter{
		Kind: "sensor", ID: "fmt", Description: "d", Mode: "computational", Event: "PostToolUse"}}})
	if !find(t, compMissing, FindingError, "run") {
		t.Errorf("expected computational sensor to require run, got %v", compMissing)
	}
}

func TestLint_DuplicateID(t *testing.T) {
	ps := []Primitive{
		{Frontmatter: Frontmatter{Kind: "guide", ID: "p/x", Description: "d"}, Path: "a.md"},
		{Frontmatter: Frontmatter{Kind: "guide", ID: "p/x", Description: "d"}, Path: "b.md"},
	}
	fs := Lint(ps)
	if !find(t, fs, FindingError, "duplicate") {
		t.Errorf("expected duplicate error, got %v", fs)
	}
}

func TestLint_SkillRequiresTriggers(t *testing.T) {
	ps := []Primitive{
		{Frontmatter: Frontmatter{Kind: "skill", ID: "demo", Description: "d"}}, // no triggers
	}
	fs := Lint(ps)
	if !find(t, fs, FindingError, "skill missing `triggers:`") {
		t.Errorf("expected skill-missing-triggers error, got %v", fs)
	}
}

func TestLint_AgentRequiresTools(t *testing.T) {
	ps := []Primitive{
		{Frontmatter: Frontmatter{Kind: "agent", ID: "security-reviewer", Description: "d"}},
	}
	fs := Lint(ps)
	if !find(t, fs, FindingError, "agent missing `tools:`") {
		t.Errorf("expected agent-missing-tools error, got %v", fs)
	}
}

func TestLint_CorpusResolve(t *testing.T) {
	// `corpus:` (was `traces:`) cites the reasoning; a dangling ref warns.
	ps := []Primitive{
		{Frontmatter: Frontmatter{Kind: "guide", ID: "p/x", Description: "d", Corpus: []string{"corpus/p/x", "corpus/missing"}}},
		{Frontmatter: Frontmatter{Kind: "corpus", ID: "corpus/p/x", Description: "d"}},
	}
	fs := Lint(ps)
	if !find(t, fs, FindingWarning, "corpus entry") {
		t.Errorf("expected unresolved-corpus warning, got %v", fs)
	}
}

// TestLint_DocumentRefsResolve — the document graph (produces/consumes/
// supersedes) is validated like corpus refs: dangling targets warn.
func TestLint_DocumentRefsResolve(t *testing.T) {
	ps := []Primitive{
		{Frontmatter: Frontmatter{Kind: "command", ID: "orient", Description: "d",
			Produces: []string{"implementation-plan"}, Consumes: []string{"missing-doc"}}},
		{Frontmatter: Frontmatter{Kind: "document", ID: "implementation-plan", Description: "d"}},
	}
	fs := Lint(ps)
	if !find(t, fs, FindingWarning, "consumes entry") {
		t.Errorf("expected unresolved-consumes warning, got %v", fs)
	}
	if find(t, fs, FindingWarning, "produces entry") {
		t.Errorf("did not expect a produces warning (target exists), got %v", fs)
	}
}

func TestLint_DescriptionTODOWarning(t *testing.T) {
	ps := []Primitive{
		{Frontmatter: Frontmatter{Kind: "command", ID: "verify", Description: "TODO — fill in"}},
	}
	fs := Lint(ps)
	if !find(t, fs, FindingWarning, "description still says TODO") {
		t.Errorf("expected TODO warning, got %v", fs)
	}
}

func TestLint_HasErrors(t *testing.T) {
	if !HasErrors([]Finding{{Severity: FindingError}}) {
		t.Error("expected HasErrors=true")
	}
	if HasErrors([]Finding{{Severity: FindingWarning}}) {
		t.Error("expected HasErrors=false with only warning")
	}
}

func TestLint_Clean(t *testing.T) {
	ps := []Primitive{
		{Frontmatter: Frontmatter{Kind: "guide", ID: "p/x", Description: "Real description."}},
		{Frontmatter: Frontmatter{Kind: "sensor", ID: "review", Description: "Real description.", Mode: "inferential", Agent: "reviewer", Returns: "findings"}},
		{Frontmatter: Frontmatter{Kind: "sensor", ID: "on-gate", Description: "Real description.", Mode: "inferential", Event: "on-gate", Agent: "reviewer", Returns: "verdict"}},
		{Frontmatter: Frontmatter{Kind: "skill", ID: "demo", Description: "Real description.", Triggers: []string{"demo"}}},
		{Frontmatter: Frontmatter{Kind: "agent", ID: "demo", Description: "Real description.", Tools: []string{"Read"}}},
		{Frontmatter: Frontmatter{Kind: "document", ID: "implementation-plan", Description: "Real description."}},
	}
	fs := Lint(ps)
	if HasErrors(fs) {
		t.Errorf("expected clean lint, got errors: %v", fs)
	}
}

func TestLint_TagsKebabCase(t *testing.T) {
	prims := []Primitive{
		{
			Frontmatter: Frontmatter{
				Kind: "guide", ID: "g1", Description: "x", Globs: []string{"x"},
				Tags: []string{"valid-tag", "Bad_Tag", "ALSO-BAD"},
			},
			Path: ".charter/guides/g1.md",
		},
	}
	findings := Lint(prims)
	errs := 0
	for _, f := range findings {
		if f.Severity == FindingError && strings.Contains(f.Message, "kebab-case") {
			errs++
		}
	}
	if errs != 2 {
		t.Errorf("expected 2 kebab-case errors, got %d (all findings: %+v)", errs, findings)
	}
}
