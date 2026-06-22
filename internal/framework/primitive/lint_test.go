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
		{Frontmatter: Frontmatter{Kind: "guide", ID: "p/x"}}, // missing description
		{Frontmatter: Frontmatter{Kind: "skill", Description: "x"}}, // missing id
		{Frontmatter: Frontmatter{ID: "x", Description: "x"}}, // missing kind
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

func TestLint_SubagentRequiresTools(t *testing.T) {
	ps := []Primitive{
		{Frontmatter: Frontmatter{Kind: "subagent", ID: "demo", Description: "d"}},
	}
	fs := Lint(ps)
	if !find(t, fs, FindingError, "subagent missing `tools:`") {
		t.Errorf("expected subagent-missing-tools error, got %v", fs)
	}
}

func TestLint_PersonaRequiresTools(t *testing.T) {
	ps := []Primitive{
		{Frontmatter: Frontmatter{Kind: "persona", ID: "security-reviewer", Description: "d"}},
	}
	fs := Lint(ps)
	if !find(t, fs, FindingError, "persona missing `tools:`") {
		t.Errorf("expected persona-missing-tools error, got %v", fs)
	}
}

func TestLint_ProjectionCollision(t *testing.T) {
	// persona id == subagent id → both project to .claude/agents/foo.md.
	ps := []Primitive{
		{Frontmatter: Frontmatter{Kind: "persona", ID: "foo", Description: "d", Tools: []string{"Read"}}, Path: "harness/personas/foo.md"},
		{Frontmatter: Frontmatter{Kind: "subagent", ID: "foo", Description: "d", Tools: []string{"Read"}}, Path: "harness/agents/foo.md"},
	}
	fs := Lint(ps)
	if !find(t, fs, FindingError, "projection collision") {
		t.Errorf("expected projection-collision error, got %v", fs)
	}
}

func TestLint_NoCollision_DifferentIds(t *testing.T) {
	ps := []Primitive{
		{Frontmatter: Frontmatter{Kind: "persona", ID: "security-reviewer", Description: "d", Tools: []string{"Read"}}, Path: "harness/personas/security-reviewer.md"},
		{Frontmatter: Frontmatter{Kind: "subagent", ID: "code-reviewer", Description: "d", Tools: []string{"Read"}}, Path: "harness/agents/code-reviewer.md"},
	}
	fs := Lint(ps)
	if HasErrors(fs) {
		t.Errorf("expected no errors with distinct ids, got %v", fs)
	}
}

func TestLint_TracesResolve(t *testing.T) {
	ps := []Primitive{
		{Frontmatter: Frontmatter{Kind: "guide", ID: "p/x", Description: "d", Traces: []string{"corpus/p/x", "corpus/missing"}}},
		{Frontmatter: Frontmatter{Kind: "corpus", ID: "corpus/p/x", Description: "d"}},
	}
	fs := Lint(ps)
	if !find(t, fs, FindingWarning, "traces entry") {
		t.Errorf("expected unresolved-trace warning, got %v", fs)
	}
}

func TestLint_DescriptionTODOWarning(t *testing.T) {
	ps := []Primitive{
		{Frontmatter: Frontmatter{Kind: "action", ID: "verify", Description: "TODO — fill in"}},
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
		{Frontmatter: Frontmatter{Kind: "skill", ID: "demo", Description: "Real description.", Triggers: []string{"demo"}}},
		{Frontmatter: Frontmatter{Kind: "subagent", ID: "demo", Description: "Real description.", Tools: []string{"Read"}}},
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
			Path: ".keystone/harness/guides/g1.md",
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
