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
		{Frontmatter: Frontmatter{Kind: "rule", ID: "p/x"}}, // missing description
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

// TestLint_RetiredKindIsUnknown — the 2.x framework vocabulary is gone in
// 3.0 with no redirect: a retired kind is simply unknown.
func TestLint_RetiredKindIsUnknown(t *testing.T) {
	for _, old := range []string{"guide", "sensor", "action", "playbook", "persona", "subagent"} {
		t.Run(old, func(t *testing.T) {
			fs := Lint([]Primitive{{Frontmatter: Frontmatter{Kind: old, ID: "x", Description: "d"}}})
			if !find(t, fs, FindingError, "unknown kind") {
				t.Errorf("kind %q: expected unknown-kind error, got %v", old, fs)
			}
		})
	}
}

func TestLint_DuplicateID(t *testing.T) {
	ps := []Primitive{
		{Frontmatter: Frontmatter{Kind: "rule", ID: "p/x", Description: "d"}, Path: "a.md"},
		{Frontmatter: Frontmatter{Kind: "rule", ID: "p/x", Description: "d"}, Path: "b.md"},
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
		{Frontmatter: Frontmatter{Kind: "rule", ID: "p/x", Description: "d", Corpus: []string{"corpus/p/x", "corpus/missing"}}},
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
		{Frontmatter: Frontmatter{Kind: "rule", ID: "p/x", Description: "Real description."}},
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
				Kind: "rule", ID: "g1", Description: "x", Globs: []string{"x"},
				Tags: []string{"valid-tag", "Bad_Tag", "ALSO-BAD"},
			},
			Path: ".keystone/harness/rules/g1.md",
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
