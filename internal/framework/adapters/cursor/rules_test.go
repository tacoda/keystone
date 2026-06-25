package cursor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

func TestProjectRules_WritesMDCPerGuide(t *testing.T) {
	root := t.TempDir()
	guide := primitive.Primitive{
		Frontmatter: primitive.Frontmatter{
			Kind:        "guide",
			ID:          "guides/idioms/go/stdlib-first",
			Description: "Prefer Go stdlib over new deps.",
			Globs:       []string{"cmd/**/*.go", "go.mod"},
		},
		Path: ".keystone/harness/guides/idioms/go/stdlib-first.md",
	}
	res, err := ProjectRules(root, []primitive.Primitive{guide})
	if err != nil {
		t.Fatalf("ProjectRules: %v", err)
	}
	if res.Wrote != 1 {
		t.Errorf("expected 1 written, got %d", res.Wrote)
	}
	data, err := os.ReadFile(filepath.Join(root, RulesDir, "go-stdlib-first.mdc"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	for _, want := range []string{
		`description: Prefer Go stdlib over new deps.`,
		`globs:`,
		`  - "cmd/**/*.go"`,
		`alwaysApply: false`,
		`source: .keystone/harness/guides/idioms/go/stdlib-first.md`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("MDC missing %q\n%s", want, s)
		}
	}
}

func TestProjectRules_SkipsNonGuides(t *testing.T) {
	root := t.TempDir()
	prims := []primitive.Primitive{
		{Frontmatter: primitive.Frontmatter{Kind: "sensor", ID: "build"}},
		{Frontmatter: primitive.Frontmatter{Kind: "guide", ID: "g", Globs: nil}}, // no globs → skip
	}
	res, err := ProjectRules(root, prims)
	if err != nil {
		t.Fatal(err)
	}
	if res.Wrote != 0 {
		t.Errorf("expected no writes, got %d", res.Wrote)
	}
}

func TestProjectRules_IsIdempotent(t *testing.T) {
	root := t.TempDir()
	p := primitive.Primitive{
		Frontmatter: primitive.Frontmatter{
			Kind: "guide", ID: "guides/idioms/go/stdlib-first",
			Description: "x", Globs: []string{"cmd/**/*.go"},
		},
		Path: ".keystone/harness/guides/idioms/go/stdlib-first.md",
	}
	if _, err := ProjectRules(root, []primitive.Primitive{p}); err != nil {
		t.Fatal(err)
	}
	res2, err := ProjectRules(root, []primitive.Primitive{p})
	if err != nil {
		t.Fatal(err)
	}
	if res2.Wrote != 0 || res2.Unchanged != 1 {
		t.Errorf("expected idempotent unchanged, got wrote=%d unchanged=%d", res2.Wrote, res2.Unchanged)
	}
}
