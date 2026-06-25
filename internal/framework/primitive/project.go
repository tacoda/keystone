package primitive

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ProjectionResult records one file written (or considered) by Project.
type ProjectionResult struct {
	Kind   string
	ID     string
	Src    string // canonical source, relative to projectDir
	Dest   string // host-native projection, relative to projectDir
	Action string // "wrote" | "skipped-unchanged" | "skipped-no-projection"
}

// Project regenerates host-native projections from canonical sources.
// For each primitive whose kind has a projection target, the canonical
// body is copied verbatim to the target path. Hand-edits at the target
// path are overwritten; the drift sensor catches them.
//
// Projection targets at 2.0:
//
//   Framework wrappers (encouraged authoring path):
//     kind: persona  → .claude/agents/<id>.md
//     kind: action   → .claude/commands/<id>.md
//     kind: playbook → .claude/skills/<id>/SKILL.md
//     kind: guide    → .claude/rules/<slug>.md  (only when globs:
//                       declared; a synthesized shim, not a verbatim
//                       copy — see writeRuleShim)
//
//   Agent escape hatches (raw host-native, same targets):
//     kind: subagent → .claude/agents/<id>.md
//     kind: command  → .claude/commands/<id>.md
//     kind: skill    → .claude/skills/<id>/SKILL.md
//
// A framework wrapper and its agent counterpart project to the same
// host path. Collisions on the same id are caught by the linter — the
// authoring layer must be unambiguous.
//
// Other kinds (corpus, sensor, eval, source, rule) and guides without
// declared globs have no host-native projection at this layer — the
// agent reads them through .keystone/INDEX.json + the canonical paths
// directly. Guides without globs are global-process content that is
// either always-on (and distilled into CLAUDE.md) or on-demand via
// INDEX.json; only globbed idiom guides need an ambient
// path-triggered shim.
//
// Disk-name normalization: ids containing `:` (canonical namespace
// separator) are rewritten to `-` for filesystem safety. The
// frontmatter id is preserved unchanged inside the projected file.
// Guide ids (slash-separated hierarchies) are flattened via
// ruleShimDiskID — see that helper for the exact transform.
func Project(projectDir string, primitives []Primitive) ([]ProjectionResult, error) {
	results := make([]ProjectionResult, 0, len(primitives))
	for _, p := range primitives {
		rel := ProjectionRelPath(p)
		if rel == "" {
			results = append(results, ProjectionResult{
				Kind: p.Kind, ID: p.ID,
				Src: p.Path, Dest: "",
				Action: "skipped-no-projection",
			})
			continue
		}
		src := filepath.Join(projectDir, p.Path)
		dest := filepath.Join(projectDir, rel)
		var writeErr error
		if Kind(p.Kind) == KindGuide {
			writeErr = writeRuleShim(src, dest, p)
		} else {
			writeErr = copyOne(src, dest)
		}
		if writeErr != nil {
			return results, fmt.Errorf("project %s/%s: %w", p.Kind, p.ID, writeErr)
		}
		results = append(results, ProjectionResult{
			Kind: p.Kind, ID: p.ID,
			Src: p.Path, Dest: rel,
			Action: "wrote",
		})
	}
	return results, nil
}

// ProjectionRelPath returns the host-native projection path for a
// primitive, relative to the project root. Returns "" for kinds that
// have no projection target.
//
// Guides project only when they declare non-empty globs — globless
// guides are global-process content with no path-triggered ambient
// channel.
func ProjectionRelPath(p Primitive) string {
	name := projectedDiskName(p.ID)
	switch Kind(p.Kind) {
	// A playbook is a composed orchestrator — it projects to a SKILL.md
	// like a skill.
	case KindSkill, KindPlaybook:
		return filepath.Join(".claude", "skills", name, "SKILL.md")
	case KindAgent:
		return filepath.Join(".claude", "agents", name+".md")
	case KindCommand:
		return filepath.Join(".claude", "commands", name+".md")
	case KindGuide:
		return guideProjection(p)
	case KindSensor:
		return sensorProjection(p, name)
	}
	// hook, pattern, posture, tool, document, corpus, concern, eval,
	// source: no host file projection here.
	return ""
}

// guideProjection returns the rule-shim path for a glob-scoped inferential
// guide (`rule` is the projection-target name). A computational guide is
// carried by a hook (no shim); a guide without globs is global-process
// content with no ambient channel.
func guideProjection(p Primitive) string {
	if p.Mode == string(modeComputational) || len(p.Globs) == 0 {
		return ""
	}
	return filepath.Join(".claude", "rules", ruleShimDiskID(p.ID)+".md")
}

// sensorProjection returns the subagent path for an inferential sensor (a
// review check dispatched as a subagent). A computational sensor runs via the
// hook layer and has no host file.
func sensorProjection(p Primitive, name string) string {
	if p.Mode == string(modeInferential) {
		return filepath.Join(".claude", "agents", name+".md")
	}
	return ""
}

// modeComputational / modeInferential mirror the `mode:` values without
// importing the validation constants into projection logic.
const (
	modeComputational = "computational"
	modeInferential   = "inferential"
)

// keystoneProjectionPrefix is prepended to every projected host artifact so
// the harness owns a clear namespace (`/keystone-<name>` for commands, etc.).
const keystoneProjectionPrefix = "keystone-"

// projectedDiskName renders a primitive id as a kebab-case, keystone-prefixed
// filesystem name for its host projection. Namespace (`:`) and hierarchy (`/`)
// separators flatten to `-`; runs collapse; an id already in the keystone
// namespace is not double-prefixed.
func projectedDiskName(id string) string {
	s := strings.ToLower(id)
	s = strings.NewReplacer(":", "-", "/", "-").Replace(s)
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	s = strings.Trim(s, "-")
	if s == "keystone" || strings.HasPrefix(s, keystoneProjectionPrefix) {
		return s
	}
	return keystoneProjectionPrefix + s
}

// ruleShimDiskID flattens a guide id into a single-segment, keystone-prefixed
// filename safe for `.claude/rules/`. The `guides/idioms/` prefix is stripped
// so the resulting filenames stay short:
//
//   guides/idioms/go/stdlib-first              → keystone-go-stdlib-first
//   guides/idioms/harness-content/state-files  → keystone-harness-content-state-files
//   guides/process/foo                         → keystone-process-foo (fallback)
func ruleShimDiskID(guideID string) string {
	trimmed := strings.TrimPrefix(guideID, "guides/idioms/")
	trimmed = strings.TrimPrefix(trimmed, "guides/")
	return projectedDiskName(trimmed)
}

func copyOne(srcAbs, destAbs string) error {
	if err := os.MkdirAll(filepath.Dir(destAbs), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(destAbs), err)
	}
	src, err := os.Open(srcAbs)
	if err != nil {
		return fmt.Errorf("open %s: %w", srcAbs, err)
	}
	defer src.Close()
	tmp, err := os.CreateTemp(filepath.Dir(destAbs), ".keystone-project.*")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := io.Copy(tmp, src); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("copy body: %w", err)
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

// writeRuleShim synthesizes a `.claude/rules/<slug>.md` file from an
// idiom guide so Claude Code's native `paths:` auto-loader fires when
// any file matches the guide's globs. The shim is intentionally
// terse — frontmatter plus the high-signal sections from the source
// (IRON LAW / GOLDEN RULE / RULES / Anti-patterns). The agent opens
// the full guide via the `source:` pointer when the rule is
// contested or it needs the prose context.
//
// Hand-edits to the shim are overwritten on the next `keystone
// project` run. Treat the keystone guide as the source of truth.
func writeRuleShim(srcAbs, destAbs string, p Primitive) error {
	data, err := os.ReadFile(srcAbs)
	if err != nil {
		return fmt.Errorf("read %s: %w", srcAbs, err)
	}
	body := stripFrontmatter(string(data))
	title := extractH1(body)
	if title == "" {
		title = ruleShimDiskID(p.ID)
	}
	sections := extractGuideSections(body)

	// Shim frontmatter mirrors the canonical primitive shape (kind /
	// id / description / globs) rather than a host-specific convention.
	// Keeps every keystone-managed file readable through the same
	// frontmatter schema. `kind: rule` signals it's an agent-escape-hatch
	// projection of the source guide. `source:` + `generated_by:`
	// document provenance so a human reading the shim isn't confused
	// about where to edit.
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("kind: rule\n")
	fmt.Fprintf(&b, "id: rules/%s\n", ruleShimDiskID(p.ID))
	b.WriteString("description: ")
	b.WriteString(yamlSafe(p.Description))
	b.WriteString("\n")
	b.WriteString("globs:\n")
	for _, g := range p.Globs {
		fmt.Fprintf(&b, "  - %q\n", g)
	}
	fmt.Fprintf(&b, "source: %s\n", p.Path)
	b.WriteString("generated_by: keystone-project\n")
	b.WriteString("---\n\n")
	fmt.Fprintf(&b, "# %s\n\n", title)
	fmt.Fprintf(&b, "Full guide: `%s` (read on demand).\n\n", p.Path)
	if sections == "" {
		b.WriteString("_No structured rule sections found in the source guide._\n")
	} else {
		b.WriteString(sections)
		if !strings.HasSuffix(sections, "\n") {
			b.WriteString("\n")
		}
	}
	return atomicWriteFile(destAbs, []byte(b.String()))
}

// stripFrontmatter returns the body of a markdown document with the
// leading YAML frontmatter block removed. If the body has no
// frontmatter, returns it unchanged. Tolerates a leading BOM or blank
// lines before the opening `---`.
func stripFrontmatter(body string) string {
	trimmed := strings.TrimLeft(body, "\ufeff \t\r\n")
	if !strings.HasPrefix(trimmed, "---") {
		return body
	}
	// Find the closing fence — a line that is exactly "---" after the opener.
	rest := trimmed[3:]
	// Require a newline after the opening fence.
	idx := strings.Index(rest, "\n")
	if idx < 0 {
		return body
	}
	rest = rest[idx+1:]
	closer := strings.Index(rest, "\n---")
	if closer < 0 {
		return body
	}
	after := rest[closer+len("\n---"):]
	// Consume the rest of the closing-fence line.
	if i := strings.Index(after, "\n"); i >= 0 {
		after = after[i+1:]
	} else {
		after = ""
	}
	return after
}

// extractH1 returns the first `# Title` line's text, without the
// leading `# `. Returns "" if no H1 is found.
func extractH1(body string) string {
	sc := bufio.NewScanner(strings.NewReader(body))
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "# ") && !strings.HasPrefix(line, "## ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}

// shimSectionAllowlist lists the H2 headings (case-insensitive,
// after `## `) that are copied verbatim into the rule shim. Other
// sections — prose explanations like "Why this is agent-specific",
// "Sensors", "See also", "Pacing modes" — stay in the source guide.
var shimSectionAllowlist = []string{
	"iron law",
	"iron laws",
	"golden rule",
	"golden rules",
	"rules",
	"anti-patterns",
	"anti patterns",
}

// extractGuideSections walks the guide body and concatenates every
// allowlisted H2 section verbatim, preserving source order. Blank
// runs are collapsed to single newlines between sections.
func extractGuideSections(body string) string {
	lines := strings.Split(body, "\n")
	var out strings.Builder
	include := false
	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			header := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(line, "## ")))
			include = false
			for _, allowed := range shimSectionAllowlist {
				if header == allowed {
					include = true
					break
				}
			}
			if include {
				if out.Len() > 0 {
					out.WriteString("\n")
				}
				out.WriteString(line)
				out.WriteString("\n")
			}
			continue
		}
		if !include {
			continue
		}
		// H1 inside the body terminates section processing.
		if strings.HasPrefix(line, "# ") && !strings.HasPrefix(line, "## ") {
			include = false
			continue
		}
		out.WriteString(line)
		out.WriteString("\n")
	}
	return strings.TrimRight(out.String(), "\n") + "\n"
}

// yamlSafe quotes a value if it contains characters that would break
// an unquoted YAML scalar. Keeps the common case (plain ASCII
// description) un-quoted for readability.
func yamlSafe(s string) string {
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

// atomicWriteFile writes contents to destAbs via a same-directory
// temp file + rename, matching copyOne's durability shape.
func atomicWriteFile(destAbs string, contents []byte) error {
	if err := os.MkdirAll(filepath.Dir(destAbs), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(destAbs), err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(destAbs), ".keystone-project.*")
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
