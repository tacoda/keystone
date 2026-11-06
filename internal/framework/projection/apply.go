package projection

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// ApplyResult reports what Apply changed, by repo-relative host path.
type ApplyResult struct {
	Wrote, Unchanged, Pruned []string
	// Owned is the new ownership list to record in the lockfile.
	Owned []string
}

// Apply materializes outs under projectDir. owned lists the host paths
// keystone created on earlier runs (lockfile `projections`). Apply never
// overwrites a path it does not own, and prunes owned paths the plan no
// longer produces.
//
// First run under 5.0 (owned is empty): 4.x projections carried a
// `keystone-` name prefix and no ownership record, so entries with that
// prefix in the directories the plan writes to are adopted as owned. That
// lets stale 4.x projections be pruned instead of left as duplicates.
func Apply(projectDir string, outs []Output, owned []string, mode LinkMode) (ApplyResult, error) {
	if len(owned) == 0 {
		owned = adoptLegacy(projectDir, outs)
	}
	isOwned := toSet(owned)
	var res ApplyResult
	for _, o := range outs {
		if o.Dest == "" {
			continue
		}
		changed, err := materialize(projectDir, o, mode, isOwned[o.Dest])
		if err != nil {
			return res, fmt.Errorf("%s %s: %w", o.Host, o.Dest, err)
		}
		if changed {
			res.Wrote = append(res.Wrote, o.Dest)
		} else {
			res.Unchanged = append(res.Unchanged, o.Dest)
		}
	}
	res.Owned = Dests(outs)
	pruned, err := prune(projectDir, owned, toSet(res.Owned))
	res.Pruned = pruned
	return res, err
}

// materialize writes one output; changed is false when disk already matches.
func materialize(projectDir string, o Output, mode LinkMode, owned bool) (bool, error) {
	dest := filepath.Join(projectDir, o.Dest)
	if !owned && exists(dest) {
		return false, fmt.Errorf("refusing to overwrite a path keystone did not create — move or remove it to let keystone project there")
	}
	switch {
	case o.Strategy == Shim:
		return writeShim(projectDir, o, dest)
	case mode == LinkCopy:
		return copyTarget(filepath.Join(projectDir, o.Target), dest)
	default:
		return linkTarget(projectDir, o, dest)
	}
}

func writeShim(projectDir string, o Output, dest string) (bool, error) {
	content, err := expectedShim(projectDir, o)
	if err != nil {
		return false, err
	}
	if prev, err := os.ReadFile(dest); err == nil && bytes.Equal(prev, content) {
		return false, nil
	}
	return true, replaceAt(dest, func() error { return os.WriteFile(dest, content, 0o644) })
}

// replaceAt clears whatever sits at dest (a file, dir, or link left by an
// older strategy), ensures its parent exists, then runs write.
func replaceAt(dest string, write func() error) error {
	if err := os.RemoveAll(dest); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	return write()
}

func expectedShim(projectDir string, o Output) ([]byte, error) {
	content, ok, err := primitive.RenderForHost(projectDir, o.Prim)
	if err == nil && !ok {
		err = fmt.Errorf("no shim renderer for %s/%s", o.Kind, o.ID)
	}
	return content, err
}

// linkTarget points dest at the canonical target with a relative symlink.
func linkTarget(projectDir string, o Output, dest string) (bool, error) {
	rel, err := relLink(projectDir, o)
	if err != nil {
		return false, err
	}
	if cur, err := os.Readlink(dest); err == nil && cur == rel {
		return false, nil
	}
	return true, replaceAt(dest, func() error { return os.Symlink(rel, dest) })
}

func relLink(projectDir string, o Output) (string, error) {
	return filepath.Rel(filepath.Dir(filepath.Join(projectDir, o.Dest)), filepath.Join(projectDir, o.Target))
}

// copyTarget mirrors the canonical file or directory at src into dest.
func copyTarget(src, dest string) (bool, error) {
	if sameTree(src, dest) {
		return false, nil
	}
	return true, replaceAt(dest, func() error { return copyTree(src, dest) })
}

func copyTree(src, dest string) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		out := filepath.Join(dest, rel)
		if d.IsDir() {
			return os.MkdirAll(out, 0o755)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return err
		}
		return os.WriteFile(out, data, 0o644)
	})
}

// sameTree reports whether dest is a real (non-link) copy of src.
func sameTree(src, dest string) bool {
	if info, err := os.Lstat(dest); err != nil || info.Mode()&os.ModeSymlink != 0 {
		return false
	}
	return treeBytes(src) == treeBytes(dest)
}

// treeBytes renders a file or directory tree as one comparable string.
func treeBytes(root string) string {
	var b strings.Builder
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(root, p)
		data, _ := os.ReadFile(p)
		fmt.Fprintf(&b, "%s\x00%s\x00", rel, data)
		return nil
	})
	return b.String()
}

// prune removes owned paths the new plan no longer produces.
func prune(projectDir string, owned []string, keep map[string]bool) ([]string, error) {
	var pruned []string
	for _, rel := range owned {
		if keep[rel] || !prunable(rel) {
			continue
		}
		abs := filepath.Join(projectDir, rel)
		if !exists(abs) {
			continue
		}
		if err := os.RemoveAll(abs); err != nil {
			return pruned, err
		}
		pruned = append(pruned, rel)
	}
	return pruned, nil
}

// prunable guards prune against a tampered or corrupt lockfile: only a
// local path (no "..", not absolute) outside the charter's own trees.
func prunable(rel string) bool {
	slash := filepath.ToSlash(rel)
	return filepath.IsLocal(rel) && !strings.HasPrefix(slash, ".agents/") && !strings.HasPrefix(slash, ".keystone/")
}

// adoptLegacy returns the `keystone-*` entries in every directory the plan
// writes into — 4.x projections, which predate ownership records.
func adoptLegacy(projectDir string, outs []Output) []string {
	var adopted []string
	seen := map[string]bool{"": true, ".": true}
	for _, o := range outs {
		dir := filepath.Dir(o.Dest)
		if !seen[dir] {
			seen[dir] = true
			adopted = append(adopted, legacyEntries(projectDir, dir)...)
		}
	}
	sort.Strings(adopted)
	return adopted
}

// legacyEntries lists repo-relative `keystone-*` entries directly under dir.
func legacyEntries(projectDir, dir string) []string {
	var out []string
	entries, _ := os.ReadDir(filepath.Join(projectDir, dir))
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "keystone-") {
			out = append(out, filepath.ToSlash(filepath.Join(dir, e.Name())))
		}
	}
	return out
}

// Drift is one host path that no longer matches the plan.
type Drift struct {
	Dest    string
	Problem string
}

// Check compares disk against the plan without writing: missing outputs,
// links pointing elsewhere, and shims or copies edited by hand.
func Check(projectDir string, outs []Output, mode LinkMode) []Drift {
	var ds []Drift
	for _, o := range outs {
		if o.Dest == "" {
			continue
		}
		if problem := checkOne(projectDir, o, mode); problem != "" {
			ds = append(ds, Drift{Dest: o.Dest, Problem: problem})
		}
	}
	return ds
}

func checkOne(projectDir string, o Output, mode LinkMode) string {
	dest := filepath.Join(projectDir, o.Dest)
	switch {
	case !exists(dest):
		return "missing — run `keystone project`"
	case o.Strategy == Shim:
		return shimDrift(projectDir, o, dest)
	case mode == LinkCopy:
		if !sameTree(filepath.Join(projectDir, o.Target), dest) {
			return "copy differs from " + o.Target
		}
	default:
		want, _ := relLink(projectDir, o)
		if cur, err := os.Readlink(dest); err != nil || cur != want {
			return "not a link to " + o.Target
		}
	}
	return ""
}

func shimDrift(projectDir string, o Output, dest string) string {
	const stale = "edited by hand or stale — edit the source and run `keystone project`"
	want, err := expectedShim(projectDir, o)
	if err != nil {
		return stale
	}
	got, err := os.ReadFile(dest)
	if err != nil || !bytes.Equal(want, got) {
		return stale
	}
	return ""
}

func exists(p string) bool {
	_, err := os.Lstat(p)
	return err == nil
}

func toSet(xs []string) map[string]bool {
	m := make(map[string]bool, len(xs))
	for _, x := range xs {
		m[x] = true
	}
	return m
}
