package primitive

import (
	"strings"
	"testing"
)

func TestCompose_NoIncludes(t *testing.T) {
	in := []Primitive{
		{Frontmatter: Frontmatter{Kind: "guide", ID: "g1", Tools: []string{"Read"}}},
	}
	out, errs := Compose(in)
	if len(errs) != 0 {
		t.Fatalf("unexpected errs: %v", errs)
	}
	if got := out[0].Tools; len(got) != 1 || got[0] != "Read" {
		t.Errorf("tools changed unexpectedly: %v", got)
	}
}

func TestCompose_UnionsListFields(t *testing.T) {
	in := []Primitive{
		{Frontmatter: Frontmatter{
			Kind: "concern", ID: "reads-diff",
			Tools: []string{"Read", "Grep"},
			Tags:  []string{"review"},
		}},
		{Frontmatter: Frontmatter{
			Kind: "persona", ID: "code-reviewer",
			Includes: []string{"reads-diff"},
			Tools:    []string{"Bash"},
			Tags:     []string{"reviewer"},
		}},
	}
	out, errs := Compose(in)
	if len(errs) != 0 {
		t.Fatalf("unexpected errs: %v", errs)
	}
	// Find the persona post-compose.
	var p Primitive
	for _, x := range out {
		if x.ID == "code-reviewer" {
			p = x
		}
	}
	wantTools := []string{"Read", "Grep", "Bash"} // concern first, host last
	if !eqSlice(p.Tools, wantTools) {
		t.Errorf("tools = %v, want %v", p.Tools, wantTools)
	}
	wantTags := []string{"review", "reviewer"}
	if !eqSlice(p.Tags, wantTags) {
		t.Errorf("tags = %v, want %v", p.Tags, wantTags)
	}
}

func TestCompose_HostWinsScalars(t *testing.T) {
	// A concern declares scalars (which is unusual but tolerated);
	// the host primitive must keep its own scalar values.
	in := []Primitive{
		{Frontmatter: Frontmatter{
			Kind: "concern", ID: "concern-with-scalars",
			Description: "concern's description",
			Severity:    "should",
			Model:       "sonnet",
		}},
		{Frontmatter: Frontmatter{
			Kind: "sensor", ID: "host-sensor",
			Description: "host's description",
			Severity:    "must",
			Model:       "opus",
			Includes:    []string{"concern-with-scalars"},
		}},
	}
	out, _ := Compose(in)
	var p Primitive
	for _, x := range out {
		if x.ID == "host-sensor" {
			p = x
		}
	}
	if p.Description != "host's description" {
		t.Errorf("description not host-wins: got %q", p.Description)
	}
	if p.Severity != "must" {
		t.Errorf("severity not host-wins: got %q", p.Severity)
	}
	if p.Model != "opus" {
		t.Errorf("model not host-wins: got %q", p.Model)
	}
}

func TestCompose_UnknownConcernIsError(t *testing.T) {
	in := []Primitive{
		{Frontmatter: Frontmatter{
			Kind: "persona", ID: "x",
			Includes: []string{"does-not-exist"},
		}},
	}
	_, errs := Compose(in)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !contains(errs[0].Message, "unknown concern") {
		t.Errorf("unexpected error: %v", errs[0])
	}
}

func TestCompose_ConcernsAreLeaves(t *testing.T) {
	in := []Primitive{
		{Frontmatter: Frontmatter{
			Kind: "concern", ID: "inner",
		}},
		{Frontmatter: Frontmatter{
			Kind: "concern", ID: "outer",
			Includes: []string{"inner"}, // illegal — concerns can't include
		}},
	}
	_, errs := Compose(in)
	if len(errs) != 1 {
		t.Fatalf("expected 1 leaf-violation error, got %d: %v", len(errs), errs)
	}
	if !contains(errs[0].Message, "leaves") {
		t.Errorf("unexpected error: %v", errs[0])
	}
}

func TestCompose_DuplicateInclude(t *testing.T) {
	in := []Primitive{
		{Frontmatter: Frontmatter{Kind: "concern", ID: "a", Tools: []string{"Read"}}},
		{Frontmatter: Frontmatter{
			Kind: "persona", ID: "host",
			Includes: []string{"a", "a"},
		}},
	}
	_, errs := Compose(in)
	if len(errs) != 1 {
		t.Fatalf("expected 1 duplicate error, got %d", len(errs))
	}
	if !contains(errs[0].Message, "duplicate") {
		t.Errorf("unexpected error: %v", errs[0])
	}
}

func TestCompose_MultiConcernUnion(t *testing.T) {
	in := []Primitive{
		{Frontmatter: Frontmatter{Kind: "concern", ID: "a", Tools: []string{"Read", "Grep"}}},
		{Frontmatter: Frontmatter{Kind: "concern", ID: "b", Tools: []string{"Bash"}}},
		{Frontmatter: Frontmatter{
			Kind: "persona", ID: "host",
			Includes: []string{"a", "b"},
			Tools:    []string{"Task"},
		}},
	}
	out, _ := Compose(in)
	var p Primitive
	for _, x := range out {
		if x.ID == "host" {
			p = x
		}
	}
	want := []string{"Read", "Grep", "Bash", "Task"}
	if !eqSlice(p.Tools, want) {
		t.Errorf("tools = %v, want %v", p.Tools, want)
	}
}

func TestCompose_HostTriggersDedup(t *testing.T) {
	// Concern + host both contribute a trigger with identical
	// (phase, matcher, command) — dedupe to one entry.
	in := []Primitive{
		{Frontmatter: Frontmatter{
			Kind: "concern", ID: "shared",
			HostTriggers: []HostTrigger{{Phase: "Stop", Command: "go test", Timeout: 30}},
		}},
		{Frontmatter: Frontmatter{
			Kind: "sensor", ID: "host",
			Includes:     []string{"shared"},
			HostTriggers: []HostTrigger{{Phase: "Stop", Command: "go test", Timeout: 30}},
		}},
	}
	out, _ := Compose(in)
	var p Primitive
	for _, x := range out {
		if x.ID == "host" {
			p = x
		}
	}
	if len(p.HostTriggers) != 1 {
		t.Errorf("expected 1 trigger after dedup, got %d: %v", len(p.HostTriggers), p.HostTriggers)
	}
}

// --- helpers ---

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

func contains(s, sub string) bool {
	return strings.Contains(s, sub)
}
