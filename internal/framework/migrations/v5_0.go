package migrations

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tacoda/keystone/internal/framework/lockfile"
)

// Version 5.0 — standard vocabulary (ADR 0010). Primitives move from
// .charter/ to .agents/ under the directory names hosts use; everything
// keystone-private moves to .keystone/. Moves only (iron law 3): no file
// content changes. 4.x frontmatter (`kind: guide`, `globs:`, `on:`) keeps
// working through parse-time normalization (iron law 6), and `keystone
// lint` names the 5.0 spelling.
//
// Paths are frozen literals, NOT config constants: a migration reproduces
// the layout of its own era and must not shift if a default moves later.
func init() {
	Register(Migration{
		Version: "5.0",
		Up:      planUp_5_0,
		Down:    planDown_5_0,
	})
}

const (
	charterRoot_5_0 = ".charter"  // 4.0 single root
	agentsRoot_5_0  = ".agents"   // 5.0 primitives
	stateRoot_5_0   = ".keystone" // 5.0 keystone-private state
)

// layout_5_0 maps each 4.x .charter/ entry to its 5.0 home, repo-relative.
// Order matters only for Down: later entries move back first.
var layout_5_0 = []struct{ from, to string }{
	{"guides", agentsRoot_5_0 + "/rules"},
	{"rules", agentsRoot_5_0 + "/rules"}, // 4.x placeholder dir; merges with guides
	{"corpus", agentsRoot_5_0 + "/references"},
	{"sensors", agentsRoot_5_0 + "/hooks"},
	{"posture", agentsRoot_5_0 + "/permissions"},
	{"evals", agentsRoot_5_0 + "/evals"},
	{"skills", agentsRoot_5_0 + "/skills"},
	{"commands", agentsRoot_5_0 + "/commands"},
	{"agents", agentsRoot_5_0 + "/agents"},
	{"tools", agentsRoot_5_0 + "/tools"},
	{"documents", agentsRoot_5_0 + "/documents"},
	{"concerns", agentsRoot_5_0 + "/concerns"}, // retired; lint warns until inlined
	{"README.md", agentsRoot_5_0 + "/README.md"},
	{"adapters", stateRoot_5_0 + "/hosts"},
	{"learning", stateRoot_5_0 + "/learning"},
	{"archive", stateRoot_5_0 + "/archive"},
	{"state", stateRoot_5_0 + "/state"},
	{"work", stateRoot_5_0 + "/work"},
	{"sources", stateRoot_5_0 + "/sources"},
	{"policies", stateRoot_5_0 + "/policies"},
	{"lockfile.json", stateRoot_5_0 + "/lockfile.json"},
	{"INDEX.json", stateRoot_5_0 + "/INDEX.json"},
	{"INDEX.lite.json", stateRoot_5_0 + "/INDEX.lite.json"},
}

// flatToSkill_5_0 are the 4.x dirs of flat `<id>.md` files that fold into
// skills: each becomes .agents/skills/<id>/SKILL.md (a rename, not a rewrite).
var flatToSkill_5_0 = []string{"playbooks", "patterns"}

func planUp_5_0(absDir string) (*Plan, error) {
	p := &Plan{}
	charterAbs := filepath.Join(absDir, charterRoot_5_0)
	if !dirExists(charterAbs) {
		return p, nil // fresh 5.0 install, or already migrated
	}
	if unknown := unknownEntries_5_0(charterAbs); len(unknown) > 0 {
		return nil, fmt.Errorf("%s/ has entries 5.0 does not know where to put: %s — move or remove them by hand, then re-run",
			charterRoot_5_0, strings.Join(unknown, ", "))
	}
	for _, dir := range flatToSkill_5_0 {
		dir := dir
		p.Add(fmt.Sprintf("fold %s/%s/<id>.md → %s/skills/<id>/SKILL.md", charterRoot_5_0, dir, agentsRoot_5_0), func(absDir string) error {
			return foldFlatToSkills_5_0(filepath.Join(absDir, charterRoot_5_0, dir), filepath.Join(absDir, agentsRoot_5_0, "skills"))
		})
	}
	for _, m := range layout_5_0 {
		m := m
		p.Add(fmt.Sprintf("move %s/%s → %s", charterRoot_5_0, m.from, m.to), func(absDir string) error {
			return mergeMove_5_0(filepath.Join(absDir, charterRoot_5_0, m.from), filepath.Join(absDir, m.to))
		})
	}
	p.Add("remove the now-empty "+charterRoot_5_0+"/", func(absDir string) error {
		return removeEmptyDir_5_0(filepath.Join(absDir, charterRoot_5_0))
	})
	p.Add("ignore "+stateRoot_5_0+"/policies/ in .gitignore where "+charterRoot_5_0+"/policies/ was ignored", func(absDir string) error {
		return carryGitignore_5_0(filepath.Join(absDir, ".gitignore"))
	})
	p.Add("record 4.x keystone-* host projections as keystone-owned (so `keystone project` may replace and prune them)", recordLegacyProjections_5_0)
	p.Add("remove 4.x generated entrypoints superseded by the AGENTS.md block (CHARTER.md; pointer-only AGENTS.md / CONVENTIONS.md)", removeLegacyEntrypoints_5_0)
	p.Add("rebuild INDEX.json + INDEX.lite.json under "+stateRoot_5_0+"/", func(absDir string) error {
		return rebuildIndex(absDir, agentsRoot_5_0, stateRoot_5_0)
	})
	return p, nil
}

// legacyEntrypoints_5_0 maps each 4.x root file to the exact prefix of the
// content 4.x generated for it. 4.x rewrote these whole files on every
// `keystone project`, so a file with this prefix holds no user content; a
// file without it is left alone.
var legacyEntrypoints_5_0 = map[string]string{
	"CHARTER.md":     "# Charter\n\nThis repo is governed by a **keystone charter**",
	"AGENTS.md":      "# Charter\n\nYou **must** read [`CHARTER.md`](CHARTER.md)",
	"CONVENTIONS.md": "# Charter\n\nYou **must** read [`CHARTER.md`](CHARTER.md)",
}

func removeLegacyEntrypoints_5_0(absDir string) error {
	for name, prefix := range legacyEntrypoints_5_0 {
		path := filepath.Join(absDir, name)
		data, err := os.ReadFile(path)
		if err != nil || !strings.HasPrefix(string(data), prefix) {
			continue // absent, or not keystone-generated
		}
		if err := os.Remove(path); err != nil {
			return err
		}
	}
	return nil
}

// legacyProjectionDirs_5_0 are the host directories 4.x projected into,
// always under a `keystone-` name.
var legacyProjectionDirs_5_0 = []string{
	".claude/skills", ".claude/agents", ".claude/commands", ".claude/rules",
	".cursor/rules", ".continue/rules",
	".opencode/skills", ".opencode/agents", ".opencode/commands", ".opencode/rules",
}

// recordLegacyProjections_5_0 adds every `keystone-*` entry in the 4.x
// projection dirs to the lockfile's `projections` list. 4.x kept no
// ownership record, and 5.0 refuses to overwrite unowned paths, so without
// this an upgraded repo would need `keystone project --overwrite`. Only
// keystone-named entries are recorded; user files are never claimed.
func recordLegacyProjections_5_0(absDir string) error {
	path := filepath.Join(absDir, stateRoot_5_0, "lockfile.json")
	lf, err := lockfile.ReadFromPath(path)
	if err != nil {
		return err
	}
	owned := map[string]bool{}
	for _, p := range lf.Projections {
		owned[p] = true
	}
	for _, dir := range legacyProjectionDirs_5_0 {
		entries, _ := os.ReadDir(filepath.Join(absDir, dir))
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), "keystone-") {
				owned[dir+"/"+e.Name()] = true
			}
		}
	}
	if len(owned) == len(lf.Projections) {
		return nil // nothing new; idempotent
	}
	lf.Projections = lf.Projections[:0]
	for p := range owned {
		lf.Projections = append(lf.Projections, p)
	}
	sort.Strings(lf.Projections)
	return lockfile.WriteToPath(path, lf)
}

// planDown_5_0 moves every 5.0 location back under .charter/. Best-effort,
// like 4.0: folded playbooks/patterns stay as skills (a folded skill is
// indistinguishable from an authored one, and 4.x reads skills/), and
// rules/ returns as guides/ (the 4.x name). The .gitignore line is kept.
func planDown_5_0(absDir string) (*Plan, error) {
	p := &Plan{}
	for i := len(layout_5_0) - 1; i >= 0; i-- {
		m := layout_5_0[i]
		if m.from == "rules" {
			continue // guides/ takes rules/ back
		}
		p.Add(fmt.Sprintf("move %s → %s/%s", m.to, charterRoot_5_0, m.from), func(absDir string) error {
			return mergeMove_5_0(filepath.Join(absDir, m.to), filepath.Join(absDir, charterRoot_5_0, m.from))
		})
	}
	for _, root := range []string{agentsRoot_5_0, stateRoot_5_0} {
		root := root
		p.Add("remove the now-empty "+root+"/ (kept if anything else lives there)", func(absDir string) error {
			_ = removeEmptyDir_5_0(filepath.Join(absDir, root))
			return nil
		})
	}
	p.Add("rebuild INDEX.json + INDEX.lite.json at the "+charterRoot_5_0+"/ root", func(absDir string) error {
		return rebuildIndex(absDir, charterRoot_5_0, charterRoot_5_0)
	})
	return p, nil
}

// unknownEntries_5_0 lists top-level .charter/ entries the 5.0 layout has
// no home for, so Up stops before moving anything (iron law 5).
func unknownEntries_5_0(charterAbs string) []string {
	known := map[string]bool{".DS_Store": true}
	for _, m := range layout_5_0 {
		known[m.from] = true
	}
	for _, d := range flatToSkill_5_0 {
		known[d] = true
	}
	entries, err := os.ReadDir(charterAbs)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if !known[e.Name()] {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out
}

// mergeMove_5_0 moves src to dst. A missing src is a no-op (idempotent).
// When dst exists, directories merge entry by entry: users may already
// keep skills under .agents/skills/ for Codex or Cursor. Two files at the
// same path are fine only when identical (a re-run); otherwise it stops
// and names the conflict.
func mergeMove_5_0(src, dst string) error {
	srcInfo, err := statOptional_5_0(src)
	if err != nil || srcInfo == nil {
		return err
	}
	dstInfo, err := statOptional_5_0(dst)
	if err != nil {
		return err
	}
	if dstInfo == nil {
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		return os.Rename(src, dst)
	}
	return mergeExisting_5_0(src, dst, srcInfo.IsDir(), dstInfo.IsDir())
}

// mergeExisting_5_0 resolves a move onto an existing dst.
func mergeExisting_5_0(src, dst string, srcDir, dstDir bool) error {
	switch {
	case srcDir && dstDir:
		return mergeDirs_5_0(src, dst)
	case srcDir || dstDir:
		return fmt.Errorf("%s and %s exist as a file and a directory — resolve by hand", src, dst)
	case sameContent_5_0(src, dst):
		return os.Remove(src)
	}
	return fmt.Errorf("both %s and %s exist with different content — resolve by hand", src, dst)
}

// statOptional_5_0 returns nil info (and no error) when path is absent.
func statOptional_5_0(path string) (os.FileInfo, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	return info, err
}

func mergeDirs_5_0(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := mergeMove_5_0(filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())); err != nil {
			return err
		}
	}
	return os.Remove(src)
}

func sameContent_5_0(a, b string) bool {
	da, errA := os.ReadFile(a)
	db, errB := os.ReadFile(b)
	return errA == nil && errB == nil && bytes.Equal(da, db)
}

// foldFlatToSkills_5_0 renames each <dir>/<id>.md to <skillsDir>/<id>/SKILL.md.
func foldFlatToSkills_5_0(dir, skillsDir string) error {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := foldOneFlat_5_0(dir, skillsDir, e); err != nil {
			return err
		}
	}
	return os.Remove(dir)
}

// foldOneFlat_5_0 folds one flat file into a skill. README.md is framework
// prose about the retired kind, so it is dropped rather than folded.
func foldOneFlat_5_0(dir, skillsDir string, e os.DirEntry) error {
	name := e.Name()
	if e.IsDir() || !strings.HasSuffix(name, ".md") {
		return fmt.Errorf("%s: unexpected entry %s — a 4.x %s/ holds flat <id>.md files only", dir, name, filepath.Base(dir))
	}
	src := filepath.Join(dir, name)
	if name == "README.md" {
		return os.Remove(src)
	}
	return mergeMove_5_0(src, filepath.Join(skillsDir, strings.TrimSuffix(name, ".md"), "SKILL.md"))
}

// removeEmptyDir_5_0 removes dir if it exists and is empty; a leftover
// entry is a conflict to surface, not to delete.
func removeEmptyDir_5_0(dir string) error {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.Name() == ".DS_Store" {
			if err := os.Remove(filepath.Join(dir, e.Name())); err != nil {
				return err
			}
			continue
		}
		return fmt.Errorf("%s is not empty after the move (%s) — resolve by hand", dir, e.Name())
	}
	return os.Remove(dir)
}

// carryGitignore_5_0 appends `.keystone/policies/` when `.gitignore`
// ignores the 4.x `.charter/policies/` (a line `keystone init` wrote), so
// vendored policies stay untracked after the move. The old line is left
// in place: editing user lines is out of bounds (iron law 2).
func carryGitignore_5_0(path string) error {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	body := string(data)
	oldLine, newLine := charterRoot_5_0+"/policies/", stateRoot_5_0+"/policies/"
	if !hasLine_5_0(body, oldLine) || hasLine_5_0(body, newLine) {
		return nil
	}
	if !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	return os.WriteFile(path, []byte(body+newLine+"\n"), 0o644)
}

func hasLine_5_0(body, line string) bool {
	for _, l := range strings.Split(body, "\n") {
		if strings.TrimSpace(l) == line {
			return true
		}
	}
	return false
}
