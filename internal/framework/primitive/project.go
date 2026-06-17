package primitive

import (
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
//   kind: skill    → .claude/skills/<id>/SKILL.md
//   kind: subagent → .claude/agents/<id>.md
//   kind: command  → .claude/commands/<id>.md
//
// Other kinds (guide, corpus, sensor, action, playbook, rule) have no
// host-native projection at this layer — the agent reads them through
// .keystone/INDEX.json + the canonical paths directly.
//
// Disk-name normalization: ids containing `:` (canonical namespace
// separator) are rewritten to `-` for filesystem safety. The
// frontmatter id is preserved unchanged inside the projected file.
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
		if err := copyOne(src, dest); err != nil {
			return results, fmt.Errorf("project %s/%s: %w", p.Kind, p.ID, err)
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
func ProjectionRelPath(p Primitive) string {
	diskID := strings.ReplaceAll(p.ID, ":", "-")
	switch Kind(p.Kind) {
	case KindSkill:
		return filepath.Join(".claude", "skills", diskID, "SKILL.md")
	case KindSubagent:
		return filepath.Join(".claude", "agents", diskID+".md")
	case KindCommand:
		return filepath.Join(".claude", "commands", diskID+".md")
	}
	return ""
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
