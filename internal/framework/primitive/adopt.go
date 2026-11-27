package primitive

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tacoda/keystone/internal/framework/config"
)

// adoptRepositoryDocs indexes repository documents as corpus entries in
// place: keystone.json's charter.roots globs (repo-relative, ** aware)
// select the files; nothing is moved or edited. Each entry gets
// provenance "repo", an id derived from its path (GLOSSARY.md →
// glossary; docs/adr/0001-naming.md → docs/adr/0001-naming), and a
// description derived from the first heading or first paragraph line.
// Canonical .agents/, .keystone/, and keystone.json never adopt.
func adoptRepositoryDocs(projectDir string) ([]Primitive, []Warning, error) {
	cfg, err := config.ReadProjectConfig(projectDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("read %s: %w", config.ProjectConfigFile, err)
	}
	if cfg == nil || cfg.Charter == nil || len(cfg.Charter.Roots) == 0 {
		return nil, nil, nil
	}

	var out []Primitive
	var warnings []Warning
	seen := map[string]bool{} // adopted id → true; one entry per id
	for _, pattern := range cfg.Charter.Roots {
		matches, err := globRepo(projectDir, filepath.ToSlash(pattern))
		if err != nil {
			return nil, nil, fmt.Errorf("charter.roots %q: %w", pattern, err)
		}
		for _, rel := range matches {
			id, desc, derr := describeAdoptedDoc(projectDir, rel)
			if derr != nil {
				warnings = append(warnings, Warning{Path: rel, Message: derr.Error()})
				continue
			}
			if seen[id] {
				warnings = append(warnings, Warning{Path: rel, Message: "adopted id " + id + " collides; not indexed twice"})
				continue
			}
			seen[id] = true
			out = append(out, Primitive{
				Frontmatter: Frontmatter{Kind: string(KindCorpus), ID: id, Description: desc},
				Path:        rel,
				Provenance:  "repo",
			})
		}
	}
	return out, warnings, nil
}

// globRepo expands one repo-relative glob against the working tree,
// returning sorted slash paths of regular files. The guard dirs (.git,
// node_modules, .agents, .keystone) never match — adopted docs come from
// the repository around the charter, not from inside it.
// GlobRepo expands one repo-relative glob against the working tree,
// returning sorted slash paths of regular files. Exported for the drift
// scanner's paths-gone check; the guard dirs (.git, node_modules,
// .agents, .keystone) never match.
func GlobRepo(projectDir, pattern string) ([]string, error) {
	return globRepo(projectDir, pattern)
}

func globRepo(projectDir, pattern string) ([]string, error) {
	// Walk the widest literal prefix; MatchPath filters per file.
	base := filepath.ToSlash(filepath.Dir(filepath.FromSlash(pattern)))
	for base != "." && base != "/" && strings.Contains(base, "*") {
		base = filepath.Dir(base)
	}
	absBase := filepath.Join(projectDir, filepath.FromSlash(base))
	info, err := os.Stat(absBase)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, nil
	}
	var out []string
	walkErr := filepath.WalkDir(absBase, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "node_modules", config.StateRoot, config.DefaultAgentsRoot:
				return fs.SkipDir
			}
			return nil
		}
		rel, relErr := filepath.Rel(projectDir, path)
		if relErr != nil {
			return relErr
		}
		rel = filepath.ToSlash(rel)
		if MatchPath(pattern, rel) {
			out = append(out, rel)
		}
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}
	sort.Strings(out)
	return out, nil
}

// describeAdoptedDoc derives an adopted file's id and description: id
// from the repo-relative path (lowercased, .md stripped), description
// from the first # heading or first non-empty, non-frontmatter line.
func describeAdoptedDoc(projectDir, rel string) (string, string, error) {
	data, err := os.ReadFile(filepath.Join(projectDir, filepath.FromSlash(rel)))
	if err != nil {
		return "", "", err
	}
	text := string(data)
	if _, body, ok := SplitFrontmatter(text); ok {
		text = body
	}
	id := strings.TrimSuffix(strings.ToLower(rel), ".md")
	return id, derivedDescription(text), nil
}

// derivedDescription returns the first # heading's text, else the first
// non-empty line, else a placeholder. The description is what the agent
// reads in the index to decide whether the body matters, so a heading
// beats nothing.
func derivedDescription(text string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "#"))
		}
		if strings.HasPrefix(line, "#") {
			continue // ## subheadings: keep waiting for the title
		}
		return line
	}
	return "(empty document)"
}
