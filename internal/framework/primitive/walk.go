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
	skillFile := func(rel string) bool {
		// Skills live at <harness-root>/skills/<id>/SKILL.md
		return filepath.Base(rel) == "SKILL.md"
	}
	evalFile := func(rel string) bool {
		// Evals live at <harness-root>/evals/<id>/EVAL.md
		return filepath.Base(rel) == "EVAL.md"
	}
	return []scanRoot{
		// Framework abstractions.
		{filepath.Join(harnessRoot, "guides"), md},
		{filepath.Join(harnessRoot, "corpus"), md},
		{filepath.Join(harnessRoot, "sensors"), md},
		{filepath.Join(harnessRoot, "actions"), md},
		{filepath.Join(harnessRoot, "playbooks"), md},
		{filepath.Join(harnessRoot, "evals"), evalFile},
		{filepath.Join(harnessRoot, "sources"), md},
		// Composition primitive — reusable mixin fragments.
		{filepath.Join(harnessRoot, "concerns"), md},
		// Agent abstractions.
		{filepath.Join(harnessRoot, "rules"), md},
		{filepath.Join(harnessRoot, "skills"), skillFile},
		{filepath.Join(harnessRoot, "agents"), md},
		{filepath.Join(harnessRoot, "commands"), md},
		{filepath.Join(harnessRoot, "personas"), md},
	}
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
