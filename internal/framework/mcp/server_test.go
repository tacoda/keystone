package mcp

import (
	"testing"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

func TestHasAllStrings(t *testing.T) {
	cases := []struct {
		name string
		have []string
		want []string
		ok   bool
	}{
		{"empty want is vacuously true", []string{"a"}, nil, true},
		{"single match", []string{"a", "b"}, []string{"a"}, true},
		{"all match", []string{"a", "b", "c"}, []string{"a", "c"}, true},
		{"missing one", []string{"a", "b"}, []string{"a", "c"}, false},
		{"empty have, non-empty want", nil, []string{"a"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := hasAllStrings(c.have, c.want)
			if got != c.ok {
				t.Errorf("hasAllStrings(%v, %v) = %v, want %v", c.have, c.want, got, c.ok)
			}
		})
	}
}

func TestBuildShowView_IncludedByReverseLookup(t *testing.T) {
	primitives := []primitive.Primitive{
		{Frontmatter: primitive.Frontmatter{Kind: "concern", ID: "reads-diff", Description: "shared review tools"}},
		{Frontmatter: primitive.Frontmatter{Kind: "persona", ID: "code-reviewer", Includes: []string{"reads-diff"}}},
		{Frontmatter: primitive.Frontmatter{Kind: "persona", ID: "security-reviewer", Includes: []string{"reads-diff"}}},
		{Frontmatter: primitive.Frontmatter{Kind: "persona", ID: "planner"}}, // does not include
	}
	view, found := buildShowView(primitives, "concern", "reads-diff")
	if !found {
		t.Fatal("expected concern to be found")
	}
	want := []string{"persona/code-reviewer", "persona/security-reviewer"}
	if !eqSlice(view.IncludedBy, want) {
		t.Errorf("IncludedBy = %v, want %v", view.IncludedBy, want)
	}
}

func TestBuildShowView_TracedByReverseLookup(t *testing.T) {
	primitives := []primitive.Primitive{
		{Frontmatter: primitive.Frontmatter{Kind: "corpus", ID: "corpus/security/owasp-top-10"}},
		{Frontmatter: primitive.Frontmatter{
			Kind: "guide", ID: "guides/process/security-review",
			Traces: []string{"corpus/security/owasp-top-10"},
		}},
	}
	view, found := buildShowView(primitives, "corpus", "corpus/security/owasp-top-10")
	if !found {
		t.Fatal("expected corpus to be found")
	}
	if len(view.TracedBy) != 1 || view.TracedBy[0] != "guide/guides/process/security-review" {
		t.Errorf("TracedBy = %v, want one guide reference", view.TracedBy)
	}
}

func TestBuildShowView_NotFound(t *testing.T) {
	primitives := []primitive.Primitive{
		{Frontmatter: primitive.Frontmatter{Kind: "guide", ID: "x"}},
	}
	_, found := buildShowView(primitives, "guide", "does-not-exist")
	if found {
		t.Error("expected not-found for missing id")
	}
}

func eqSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
