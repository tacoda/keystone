package primitive

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// scanRoot is a single directory the walker descends and the file
// pattern it accepts. Files outside these roots are not indexed.
type scanRoot struct {
	// dir is repo-relative POSIX. The walker joins it with projectDir.
	dir string
	// match returns true for files (relative to dir) that should be
	// fed to the frontmatter parser. README.md files are filtered here.
	match func(rel string) bool
}

// scanRoots returns the directories the indexer descends, given the
// configured harness root path. Every canonical primitive lives under
// <harness-root>/ at 2.0 — the host-native `.claude/` tree is a
// regenerated *projection* of these sources, not a source itself, so
// the indexer never walks it.
func scanRoots(harnessRoot string) []scanRoot {
	md := func(rel string) bool {
		base := filepath.Base(rel)
		if base == "README.md" {
			return false
		}
		return strings.HasSuffix(rel, ".md")
	}
	// Skills live at <harness-root>/skills/<id>/SKILL.md; evals at
	// <harness-root>/evals/<id>/EVAL.md. Everything else is a flat .md.
	matchFor := func(dir string) func(string) bool {
		switch dir {
		case "skills":
			return func(rel string) bool { return filepath.Base(rel) == "SKILL.md" }
		case "evals":
			return func(rel string) bool { return filepath.Base(rel) == "EVAL.md" }
		default:
			return md
		}
	}
	// Derive the scan set from canonicalDirKind so the walker and the
	// kind taxonomy can never drift (the bug that hid hooks/patterns/
	// posture/tools/documents from the index). Sorted for deterministic
	// walk order.
	dirs := make([]string, 0, len(canonicalDirKind))
	for dir := range canonicalDirKind {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)
	roots := make([]scanRoot, 0, len(dirs))
	for _, dir := range dirs {
		roots = append(roots, scanRoot{filepath.Join(harnessRoot, dir), matchFor(dir)})
	}
	return roots
}

// Walk scans every configured root under projectDir, parses each
// primitive file's frontmatter, and returns the descriptors plus any
// parse warnings (files with frontmatter that failed to decode).
//
// Files without frontmatter are silently skipped — the migration step
// will land canonical frontmatter on every harness file, but the
// indexer must work on partial installs too.
func Walk(projectDir, harnessRoot string) (primitives []Primitive, warnings []Warning, err error) {
	for _, root := range scanRoots(harnessRoot) {
		abs := filepath.Join(projectDir, root.dir)
		info, statErr := os.Stat(abs)
		if statErr != nil {
			if os.IsNotExist(statErr) {
				continue
			}
			return nil, nil, fmt.Errorf("stat %s: %w", root.dir, statErr)
		}
		if !info.IsDir() {
			continue
		}
		walkErr := filepath.WalkDir(abs, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() {
				if isWorkspaceDir(abs, path) {
					return fs.SkipDir
				}
				return nil
			}
			rel, relErr := filepath.Rel(abs, path)
			if relErr != nil {
				return relErr
			}
			if !root.match(filepath.ToSlash(rel)) {
				return nil
			}
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return fmt.Errorf("read %s: %w", path, readErr)
			}
			fm, ok, parseErr := Parse(string(data))
			if parseErr != nil {
				warnings = append(warnings, Warning{
					Path:    relProject(projectDir, path),
					Message: parseErr.Error(),
				})
				return nil
			}
			if !ok {
				// No frontmatter — file is pre-migration. Skip silently.
				return nil
			}
			relPath := filepath.ToSlash(relProject(projectDir, path))
			// Convention over configuration: an omitted `kind:` is
			// inferred from the canonical directory. Explicit `kind:`
			// wins (escape hatch).
			// Convention over configuration: an omitted kind is inferred
			// from the canonical directory; an explicit kind wins.
			fm.Kind = resolveKind(fm.Kind, relPath)
			primitives = append(primitives, Primitive{
				Frontmatter: fm,
				Path:        relPath,
				Provenance:  derivProvenance(relPath, harnessRoot),
			})
			return nil
		})
		if walkErr != nil {
			return nil, nil, walkErr
		}
	}
	sort.Slice(primitives, func(i, j int) bool {
		if primitives[i].Kind != primitives[j].Kind {
			return primitives[i].Kind < primitives[j].Kind
		}
		return primitives[i].ID < primitives[j].ID
	})
	return primitives, warnings, nil
}

// isWorkspaceDir reports whether path is the top-level `work/` directory
// under the scan root. The document workspace holds filled document
// *instances* (gate state), not primitive templates, so the walker skips
// it — instances are tracked by `keystone document`, not the index.
func isWorkspaceDir(scanRootAbs, path string) bool {
	r, err := filepath.Rel(scanRootAbs, path)
	return err == nil && filepath.ToSlash(r) == "work"
}

// Warning is a non-fatal indexer finding — e.g. malformed frontmatter
// in a single file. The caller prints these to stderr; the index still
// emits with the surviving primitives.
type Warning struct {
	Path    string
	Message string
}

// derivProvenance turns a primitive's repo-relative path into the
// cascade layer it ships under. Paths inside `<harness>/policies/<name>/`
// are policy-vendored; anything else under `<harness>/` is project.
//
// Examples (with harnessRoot = ".keystone/harness"):
//   .keystone/harness/guides/process/spec.md            → "project"
//   .keystone/harness/policies/acme/guides/x.md         → "policy/acme"
//   .keystone/harness/policies/acme/nested/policies/b/  → "policy/acme" (outermost wins)
func derivProvenance(relPath, harnessRoot string) string {
	policiesPrefix := filepath.ToSlash(filepath.Join(harnessRoot, "policies")) + "/"
	rel := filepath.ToSlash(relPath)
	if !strings.HasPrefix(rel, policiesPrefix) {
		return "project"
	}
	rest := strings.TrimPrefix(rel, policiesPrefix)
	if i := strings.IndexByte(rest, '/'); i >= 0 {
		return "policy/" + rest[:i]
	}
	return "policy/" + rest
}

func relProject(projectDir, path string) string {
	if rel, err := filepath.Rel(projectDir, path); err == nil {
		return rel
	}
	return path
}
