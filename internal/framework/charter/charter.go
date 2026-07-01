// Package charter derives read-side views over a project's charter that
// the CLI, the web dashboard, and the MCP server all share: coverage
// (which files a guide governs) and the effective post-cascade roster.
// Keeping the logic here means the three surfaces never drift.
package charter

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// CoverageResult is the outcome of a coverage scan: how many project
// files a guide governs vs. how many are uncharted (matched by none).
type CoverageResult struct {
	Total     int
	Governed  int
	Uncharted []string // project-relative POSIX paths, no guide matches
}

// UnchartedByRegion groups the uncharted paths by their top-level
// directory (or "(root)"), for a compact report.
func (r CoverageResult) UnchartedByRegion() map[string]int {
	counts := map[string]int{}
	for _, rel := range r.Uncharted {
		counts[topSegment(rel)]++
	}
	return counts
}

// Coverage walks the project's source files and classifies each as
// governed (matched by ≥1 guide glob) or uncharted. Generated,
// vendored, and hidden trees are skipped — coverage is about the source
// an agent edits.
func Coverage(projectDir, charterRoot string) (CoverageResult, error) {
	globs, err := guideGlobs(projectDir, charterRoot)
	if err != nil {
		return CoverageResult{}, err
	}
	sc := &covScan{projectDir: projectDir, globs: globs}
	if err := filepath.WalkDir(projectDir, sc.visit); err != nil {
		return CoverageResult{}, err
	}
	return sc.res, nil
}

// covScan carries the walk state so visit matches WalkDir's signature.
type covScan struct {
	projectDir string
	globs      []string
	res        CoverageResult
}

// visit classifies one WalkDir entry into the running result.
func (sc *covScan) visit(path string, d fs.DirEntry, walkErr error) error {
	if walkErr != nil {
		return nil
	}
	if d.IsDir() {
		if path != sc.projectDir && skipDir(d.Name()) {
			return filepath.SkipDir
		}
		return nil
	}
	rel, e := filepath.Rel(sc.projectDir, path)
	if e != nil {
		return nil
	}
	rel = filepath.ToSlash(rel)
	sc.res.Total++
	if anyMatch(sc.globs, rel) {
		sc.res.Governed++
	} else {
		sc.res.Uncharted = append(sc.res.Uncharted, rel)
	}
	return nil
}

// guideGlobs returns every positive glob any guide claims (post-cascade).
func guideGlobs(projectDir, charterRoot string) ([]string, error) {
	prims, _, err := primitive.Walk(projectDir, charterRoot)
	if err != nil {
		return nil, err
	}
	composed, _ := primitive.Compose(prims)
	var globs []string
	for _, p := range composed {
		if primitive.Kind(p.Kind) == primitive.KindGuide {
			globs = append(globs, positiveGlobs(p.Globs)...)
		}
	}
	return globs, nil
}

// positiveGlobs drops `!`-negated patterns — a negation excludes files,
// it doesn't grant coverage.
func positiveGlobs(gs []string) []string {
	var out []string
	for _, g := range gs {
		if !strings.HasPrefix(g, "!") {
			out = append(out, g)
		}
	}
	return out
}

func anyMatch(globs []string, rel string) bool {
	for _, g := range globs {
		if primitive.MatchPath(g, rel) {
			return true
		}
	}
	return false
}

func skipDir(name string) bool {
	if strings.HasPrefix(name, ".") {
		return true
	}
	switch name {
	case "node_modules", "vendor", "dist", "build":
		return true
	}
	return false
}

func topSegment(rel string) string {
	if i := strings.Index(rel, "/"); i >= 0 {
		return rel[:i] + "/"
	}
	return "(root)"
}

// Entry is one line of the effective roster: the winning primitive plus
// the cascade layers it shadows (empty unless an override occurred).
type Entry struct {
	Primitive primitive.Primitive
	Shadows   []string
}

// Effective collapses same-id primitives to their cascade winner
// (project over policy; first-seen among policies), preserving order and
// recording what each winner shadows.
func Effective(prims []primitive.Primitive) []Entry {
	byID := map[string]*Entry{}
	var order []string
	for _, p := range prims {
		key := p.Kind + "/" + p.ID
		if e, ok := byID[key]; ok {
			mergeCascade(e, p)
			continue
		}
		byID[key] = &Entry{Primitive: p}
		order = append(order, key)
	}
	out := make([]Entry, 0, len(order))
	for _, k := range order {
		out = append(out, *byID[k])
	}
	return out
}

func mergeCascade(e *Entry, p primitive.Primitive) {
	if e.Primitive.Provenance != "project" && p.Provenance == "project" {
		e.Shadows = append(e.Shadows, e.Primitive.Provenance)
		e.Primitive = p
		return
	}
	e.Shadows = append(e.Shadows, p.Provenance)
}

// Signals returns the known signals for a project: keystone's built-ins
// plus any custom names, sorted and de-duplicated.
func Signals(custom []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range append(append([]string{}, primitive.BuiltinSignals...), custom...) {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}
