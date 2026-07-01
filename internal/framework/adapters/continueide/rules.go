// Package continueide projects keystone's idiom guides into the
// Continue IDE extension's rule layout. Package name is `continueide`
// (not `continue` — `continue` is a Go keyword and would conflict).
//
// Continue reads `.continue/rules/*.md` files in newer versions; older
// versions consult `.continue/config.json` rules arrays. The projector
// emits the modern `.continue/rules/<slug>.md` shape — users on older
// Continue can copy the bodies into their config manually.
package continueide

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tacoda/keystone/internal/framework/adapters/agnostic"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// RulesDir is the directory Continue watches for project rules.
const RulesDir = ".continue/rules"

// ProjectionResult records files written / unchanged.
type ProjectionResult struct {
	Wrote     int
	Unchanged int
}

// tally records one file outcome.
func (r *ProjectionResult) tally(wrote bool) {
	if wrote {
		r.Wrote++
	} else {
		r.Unchanged++
	}
}

// writeCharterPointer writes the always-on charter pointer rule,
// reporting whether the file changed.
func writeCharterPointer(projectDir string) (bool, error) {
	dest := filepath.Join(projectDir, RulesDir, "keystone-charter.md")
	content := buildCharterRule()
	if prev, _ := os.ReadFile(dest); bytes.Equal(prev, content) {
		return false, nil
	}
	if err := atomicWrite(dest, content); err != nil {
		return false, fmt.Errorf("write %s: %w", dest, err)
	}
	return true, nil
}

// ProjectRules writes one `.continue/rules/<slug>.md` per idiom guide
// with non-empty globs. Continue rules are simpler than Cursor's MDC
// — plain markdown with optional `---` frontmatter naming the rule
// and its applicability. The projector emits the frontmatter Continue
// understands plus the shim body.
func ProjectRules(projectDir string, primitives []primitive.Primitive) (ProjectionResult, error) {
	var out ProjectionResult

	// Always-on charter pointer: directs the agent to CHARTER.md (iron
	// laws + ambient rules) regardless of which file is in focus.
	wrote, err := writeCharterPointer(projectDir)
	if err != nil {
		return out, err
	}
	out.tally(wrote)

	for _, p := range primitives {
		if primitive.Kind(p.Kind) != primitive.KindGuide || len(p.Globs) == 0 {
			continue
		}
		slug := ruleSlug(p.ID)
		dest := filepath.Join(projectDir, RulesDir, slug+".md")
		content := buildContinueRule(p)
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

func ruleSlug(guideID string) string {
	trimmed := strings.TrimPrefix(guideID, "guides/idioms/")
	trimmed = strings.TrimPrefix(trimmed, "guides/")
	slug := strings.ReplaceAll(trimmed, "/", "-")
	if !strings.HasPrefix(slug, "keystone-") {
		slug = "keystone-" + slug
	}
	return slug
}

func buildContinueRule(p primitive.Primitive) []byte {
	var b bytes.Buffer
	b.WriteString("---\n")
	fmt.Fprintf(&b, "name: %s\n", ruleSlug(p.ID))
	fmt.Fprintf(&b, "description: %s\n", yamlScalar(p.Description))
	b.WriteString("globs:\n")
	for _, g := range p.Globs {
		fmt.Fprintf(&b, "  - %q\n", g)
	}
	fmt.Fprintf(&b, "source: %s\n", p.Path)
	b.WriteString("generated_by: keystone-project\n")
	b.WriteString("---\n\n")
	fmt.Fprintf(&b, "Full guide: `%s` (read on demand).\n\n", p.Path)
	b.WriteString("This rule auto-applies in Continue when a touched file matches the globs above. The body below is the high-signal subset (iron law + golden rule + rules + anti-patterns); open the source guide for prose context.\n")
	return b.Bytes()
}

// buildCharterRule composes the always-on charter pointer rule for
// Continue. Body is the shared thin pointer; no globs so it applies
// broadly.
func buildCharterRule() []byte {
	var b bytes.Buffer
	b.WriteString("---\n")
	b.WriteString("name: keystone-charter\n")
	b.WriteString("description: Keystone charter entrypoint — read CHARTER.md first\n")
	b.WriteString("generated_by: keystone-project\n")
	b.WriteString("---\n\n")
	b.WriteString(agnostic.RenderPointer(agnostic.ContinueProfile()))
	return b.Bytes()
}

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

func atomicWrite(destAbs string, contents []byte) error {
	if err := os.MkdirAll(filepath.Dir(destAbs), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(destAbs), err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(destAbs), ".keystone-continue.*")
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
