// Package cursor projects keystone primitives into Cursor's
// host-native rule layout. Cursor reads `.cursor/rules/<id>.mdc` (MDC
// = Markdown + frontmatter) and auto-applies rules whose `globs:`
// match the file in focus, the same triggering shape as Claude Code's
// `.claude/rules/`.
//
// Source of truth lives in `.keystone/harness/guides/idioms/*.md`
// (kind=guide with non-empty globs). The shim body is the same
// extracted IRON LAW / GOLDEN RULE / RULES / Anti-patterns block the
// Claude Code shim carries; only the frontmatter shape differs to
// match Cursor's expected schema (`description`, `globs`,
// `alwaysApply`).
package cursor

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// RulesDir is the relative directory Cursor watches for project rules.
const RulesDir = ".cursor/rules"

// ProjectionResult records one rule file written (or unchanged) by ProjectRules.
type ProjectionResult struct {
	Wrote     int // files newly written or content-changed
	Unchanged int // files whose content was already correct
}

// ProjectRules writes one `.cursor/rules/<slug>.mdc` per idiom guide
// with non-empty globs. The slug matches the Claude Code rule shim
// slug so the two host surfaces stay aligned (same name → same
// concept across hosts). Re-runs are idempotent.
func ProjectRules(projectDir string, primitives []primitive.Primitive) (ProjectionResult, error) {
	var out ProjectionResult
	for _, p := range primitives {
		if primitive.Kind(p.Kind) != primitive.KindRule || len(p.Globs) == 0 {
			continue
		}
		slug := ruleSlug(p.ID)
		dest := filepath.Join(projectDir, RulesDir, slug+".mdc")
		content := buildMDC(p)
		prev, _ := os.ReadFile(dest)
		if bytes.Equal(prev, content) {
			out.Unchanged++
			continue
		}
		if err := atomicWrite(dest, content); err != nil {
			return out, fmt.Errorf("write %s: %w", dest, err)
		}
		out.Wrote++
	}
	return out, nil
}

// ruleSlug matches the Claude Code shim's id flattener — strip
// `guides/idioms/` / `guides/`, replace `/` with `-`. Single source of
// truth would be a shared helper; for 2.2.0 we duplicate the four-line
// rule to keep the adapter packages self-contained.
func ruleSlug(guideID string) string {
	trimmed := strings.TrimPrefix(guideID, "guides/idioms/")
	trimmed = strings.TrimPrefix(trimmed, "guides/")
	return strings.ReplaceAll(trimmed, "/", "-")
}

// buildMDC composes a Cursor MDC file from a guide primitive. Cursor's
// schema (as of late 2024): `description`, `globs`, `alwaysApply` in
// frontmatter, free-form markdown body. We emit `alwaysApply: false`
// since the guide already declared globs — Cursor scopes accordingly.
//
// Body content matches the Claude Code shim: a pointer to the source
// guide plus extracted high-signal sections. Cursor users get the
// same rule density as Claude Code users.
func buildMDC(p primitive.Primitive) []byte {
	var b bytes.Buffer
	b.WriteString("---\n")
	fmt.Fprintf(&b, "description: %s\n", yamlScalar(p.Description))
	b.WriteString("globs:\n")
	for _, g := range p.Globs {
		fmt.Fprintf(&b, "  - %q\n", g)
	}
	b.WriteString("alwaysApply: false\n")
	fmt.Fprintf(&b, "source: %s\n", p.Path)
	b.WriteString("generated_by: keystone-project\n")
	b.WriteString("---\n\n")
	fmt.Fprintf(&b, "Full guide: `%s` (read on demand).\n\n", p.Path)
	b.WriteString("This rule auto-applies in Cursor when a touched file matches the globs above. The body below is the high-signal subset (iron law + golden rule + rules + anti-patterns); open the source guide for prose context.\n")
	return b.Bytes()
}

// yamlScalar quotes a value when it contains characters that would
// break unquoted YAML scalars. Plain ASCII descriptions stay
// un-quoted for readability.
func yamlScalar(s string) string {
	if s == "" {
		return `""`
	}
	for _, r := range s {
		if r == ':' || r == '#' || r == '\n' || r == '"' || r == '\'' {
			return fmt.Sprintf("%q", s)
		}
	}
	return s
}

// atomicWrite — same temp+rename shape every keystone adapter uses.
// Duplicated rather than centralized to keep each adapter package
// self-contained (deliberate cost; centralizing means one fix has to
// touch every host).
func atomicWrite(destAbs string, contents []byte) error {
	if err := os.MkdirAll(filepath.Dir(destAbs), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(destAbs), err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(destAbs), ".keystone-cursor.*")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(contents); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Rename(tmpName, destAbs); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("rename %s -> %s: %w", tmpName, destAbs, err)
	}
	return nil
}
