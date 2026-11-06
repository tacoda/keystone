// Package allocate converts an existing agent setup — host-specific files
// such as .claude/, .cursor/rules, Copilot instructions, Cline rules — into
// the charter's .agents/ sources. It never changes an original: Plan reports
// where each item would go, and Apply writes only new files under .agents/.
package allocate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// Item is one piece of an existing setup and where it goes in .agents/.
type Item struct {
	Source  string         // project-relative source path
	Kind    primitive.Kind // target kind; "" means it stays where it is
	Target  string         // project-relative .agents/ path; "" when it stays
	Content []byte         // the converted file
	Note    string         // what changed in conversion, or why it stays
	Tokens  int            // always-loaded tokens the move saves (root-file sections)

	// What --overwrite hands to keystone once the target is written; see
	// Overwrite. At most one is set.
	own     string     // a host path keystone takes ownership of
	section string     // root-file text to remove (root file is Source's file)
	entries []string   // settings.json entries ("bucket:rule") keystone manages
	hook    *movedHook // a settings.json hook to remove
}

// Moves reports whether the item has a target in .agents/.
func (it Item) Moves() bool { return it.Target != "" }

// Plan scans every known host-specific location under projectDir. owned
// lists the paths keystone itself projects (lockfile `projections`); those
// and anything reached through a symlink are keystone's output, not input.
// managed is the lockfile's `managed_entries` (keystone's own entries in
// shared files).
func Plan(projectDir string, owned []string, managed map[string][]string) ([]Item, error) {
	s := scan{dir: projectDir, owned: ownedSet(owned), managed: managed}
	for _, src := range sources {
		if err := src(&s); err != nil {
			return nil, err
		}
	}
	s.markDuplicates()
	sort.SliceStable(s.items, func(i, j int) bool { return s.items[i].Source < s.items[j].Source })
	return s.items, nil
}

// Result is what Apply did: files written, and items skipped because the
// target already exists.
type Result struct {
	Wrote, Exists []string
}

// Apply writes each moving item's content to its target. An existing
// target is never overwritten; it is listed in Exists. Sources are never
// touched.
func Apply(projectDir string, items []Item) (Result, error) {
	var res Result
	for _, it := range items {
		if !it.Moves() {
			continue
		}
		dest := filepath.Join(projectDir, it.Target)
		if _, err := os.Lstat(dest); err == nil {
			res.Exists = append(res.Exists, it.Target)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return res, err
		}
		if err := os.WriteFile(dest, it.Content, 0o644); err != nil {
			return res, fmt.Errorf("write %s: %w", it.Target, err)
		}
		res.Wrote = append(res.Wrote, it.Target)
	}
	return res, nil
}

// scan accumulates items while the sources run.
type scan struct {
	dir     string
	owned   map[string]bool
	managed map[string][]string
	items   []Item
}

func (s *scan) add(it Item) { s.items = append(s.items, it) }

// markDuplicates keeps the first item per target; a later one (the same
// skill under .claude/ and .cursor/, say) stays where it is.
func (s *scan) markDuplicates() {
	seen := map[string]string{}
	for i, it := range s.items {
		if !it.Moves() {
			continue
		}
		if first, dup := seen[it.Target]; dup {
			s.items[i] = Item{Source: it.Source, Note: "same target as " + first + "; compare by hand"}
			continue
		}
		seen[it.Target] = it.Source
	}
}

// files lists the regular files under rel (project-relative) whose names
// end in suffix, skipping keystone's own output. A missing dir is empty.
func (s *scan) files(rel, suffix string) []string {
	root := filepath.Join(s.dir, rel)
	var out []string
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.Type()&os.ModeSymlink != 0 {
			return nil // a symlink is a keystone link (or not ours to convert)
		}
		r, _ := filepath.Rel(s.dir, path)
		r = filepath.ToSlash(r)
		if s.isOwned(r) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), suffix) {
			out = append(out, r)
		}
		return nil
	})
	return out
}

// file returns rel when it is a regular, unowned file.
func (s *scan) file(rel string) (string, bool) {
	info, err := os.Lstat(filepath.Join(s.dir, rel))
	if err != nil || !info.Mode().IsRegular() {
		return "", false
	}
	return rel, !s.isOwned(rel)
}

func (s *scan) read(rel string) []byte {
	data, _ := os.ReadFile(filepath.Join(s.dir, rel))
	return data
}

// isOwned: rel or one of its parent dirs is a keystone projection.
func (s *scan) isOwned(rel string) bool {
	for p := rel; p != "." && p != "/"; p = filepath.ToSlash(filepath.Dir(p)) {
		if s.owned[p] {
			return true
		}
	}
	return false
}

func ownedSet(owned []string) map[string]bool {
	set := map[string]bool{}
	for _, o := range owned {
		set[filepath.ToSlash(o)] = true
	}
	return set
}
