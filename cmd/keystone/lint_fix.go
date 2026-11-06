package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// fixFrontmatter applies the 5.0 frontmatter fixes to project primitives:
// SKILL.md files move into Agent Skills spec shape; every other primitive
// gets 5.0 spelling (`paths:`, `event:`, 5.0 kind names) and its retired
// concerns inlined. Bodies are untouched, vendored policy files are skipped
// (read-only), and each rewrite prints its new frontmatter, so nothing is
// overwritten silently.
func fixFrontmatter(absDir string, primitives []primitive.Primitive) (int, error) {
	composed, _ := primitive.Compose(primitives)
	fixed := 0
	for _, p := range composed {
		if p.Provenance != "project" {
			continue
		}
		abs := filepath.Join(absDir, p.Path)
		fm, err := fixedFrontmatter(abs, p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  skip: %v\n", err)
			continue
		}
		if fm == "" {
			continue
		}
		if err := rewriteFrontmatter(abs, fm); err != nil {
			return fixed, err
		}
		fmt.Fprintf(os.Stdout, "  fixed: %s\n    %s\n", p.Path, strings.ReplaceAll(strings.TrimSpace(fm), "\n", "\n    "))
		fixed++
	}
	return fixed, nil
}

// fixedFrontmatter returns p's rewritten frontmatter, or "" when p needs no fix.
func fixedFrontmatter(abs string, p primitive.Primitive) (string, error) {
	if primitive.NeedsSpecFix(p) {
		return primitive.SpecSkillFrontmatter(p)
	}
	if !primitive.NeedsModernize(p) {
		return "", nil
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return "", err
	}
	raw, _, _ := primitive.SplitFrontmatter(string(data))
	return primitive.ModernizeFrontmatter(raw, p)
}

// rewriteFrontmatter replaces the frontmatter block of the file at abs,
// keeping the body byte-for-byte.
func rewriteFrontmatter(abs, fm string) error {
	data, err := os.ReadFile(abs)
	if err != nil {
		return err
	}
	_, body, ok := primitive.SplitFrontmatter(string(data))
	if !ok {
		return fmt.Errorf("%s: no frontmatter to rewrite", abs)
	}
	return os.WriteFile(abs, []byte("---\n"+fm+"---\n"+body), 0o644)
}
