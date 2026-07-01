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
		Path: ".charter/guides/idioms/go/stdlib-first.md",
	}
	res, err := ProjectRules(root, []primitive.Primitive{guide})
	if err != nil {
		t.Fatalf("ProjectRules: %v", err)
	}
	// One per-guide rule + the always-apply charter pointer.
	if res.Wrote != 2 {
		t.Errorf("expected 2 written (guide + charter pointer), got %d", res.Wrote)
	}
	data, err := os.ReadFile(filepath.Join(root, RulesDir, "keystone-go-stdlib-first.mdc"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	for _, want := range []string{
		`description: Prefer Go stdlib over new deps.`,
		`globs:`,
		`  - "cmd/**/*.go"`,
		`alwaysApply: false`,
		`source: .charter/guides/idioms/go/stdlib-first.md`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("MDC missing %q\n%s", want, s)
		}
	}
}

func TestProjectRules_WritesCharterPointer(t *testing.T) {
	root := t.TempDir()
	if _, err := ProjectRules(root, nil); err != nil {
		t.Fatalf("ProjectRules: %v", err)
	}
	charter, err := os.ReadFile(filepath.Join(root, RulesDir, "keystone-charter.mdc"))
	if err != nil {
		t.Fatalf("charter pointer not written: %v", err)
	}
	for _, want := range []string{"alwaysApply: true", "CHARTER.md"} {
		if !strings.Contains(string(charter), want) {
			t.Errorf("charter pointer missing %q\n%s", want, charter)
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
	// Only the always-apply charter pointer is written; the sensor and
	// the globless guide are both skipped.
	if res.Wrote != 1 {
		t.Errorf("expected only the charter pointer written, got %d", res.Wrote)
	}
	if _, err := os.Stat(filepath.Join(root, RulesDir, "keystone-charter.mdc")); err != nil {
		t.Errorf("charter pointer should still be written: %v", err)
	}
}

func TestProjectRules_IsIdempotent(t *testing.T) {
	root := t.TempDir()
	p := primitive.Primitive{
		Frontmatter: primitive.Frontmatter{
			Kind: "guide", ID: "guides/idioms/go/stdlib-first",
			Description: "x", Globs: []string{"cmd/**/*.go"},
		},
		Path: ".charter/guides/idioms/go/stdlib-first.md",
	}
	if _, err := ProjectRules(root, []primitive.Primitive{p}); err != nil {
		t.Fatal(err)
	}
	res2, err := ProjectRules(root, []primitive.Primitive{p})
	if err != nil {
		t.Fatal(err)
	}
	// Guide rule + charter pointer both unchanged on the second run.
	if res2.Wrote != 0 || res2.Unchanged != 2 {
		t.Errorf("expected idempotent unchanged, got wrote=%d unchanged=%d", res2.Wrote, res2.Unchanged)
	}
}
