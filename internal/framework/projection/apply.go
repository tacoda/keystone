package projection

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// ApplyResult reports what Apply changed, by repo-relative host path.
type ApplyResult struct {
	Wrote, Unchanged, Pruned []string
	// Owned is the new ownership list to record in the lockfile.
	Owned []string
}

// Options controls Apply.
type Options struct {
	Mode LinkMode
	// Overwrite lets Apply replace a path keystone does not own. Off by
	// default: keystone never overwrites what it did not create unless the
	// user passes --overwrite.
	Overwrite bool
}

// Apply materializes outs under projectDir. owned lists the host paths
// keystone created on earlier runs (lockfile `projections`). Apply refuses
// to overwrite a path it does not own (unless opts.Overwrite), and prunes
// owned paths the plan no longer produces.
func Apply(projectDir string, outs []Output, owned []string, opts Options) (ApplyResult, error) {
	isOwned := toSet(owned)
	var res ApplyResult
	if err := preflight(projectDir, outs, isOwned, opts); err != nil {
		return res, err
	}
	for _, o := range outs {
		if o.Dest == "" {
			continue
		}
		changed, err := materialize(projectDir, o, opts)
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

// preflight refuses the whole plan, before anything is written, when any
// output would replace a path keystone does not own — so a refused run
// never leaves a half-applied projection behind. It names every conflict.
func preflight(projectDir string, outs []Output, owned map[string]bool, opts Options) error {
	var conflicts []string
	for _, o := range outs {
		if o.Dest != "" && blocked(filepath.Join(projectDir, o.Dest), owned[o.Dest], opts) {
			conflicts = append(conflicts, o.Dest)
		}
	}
	if len(conflicts) == 0 {
		return nil
	}
	return fmt.Errorf("refusing to overwrite %d path(s) keystone did not create — move them, or re-run with --overwrite:\n  %s",
		len(conflicts), strings.Join(conflicts, "\n  "))
}

// materialize writes one output; changed is false when disk already matches.
func materialize(projectDir string, o Output, opts Options) (bool, error) {
	dest := filepath.Join(projectDir, o.Dest)
	switch {
	case o.Strategy == Shim:
		return writeShim(projectDir, o, dest)
	case o.Strategy == Copy || opts.Mode == LinkCopy:
		return copyTarget(filepath.Join(projectDir, o.Target), dest)
	default:
		return linkTarget(projectDir, o, dest)
	}
}

// blocked reports whether dest holds something keystone may not replace:
// an existing path it does not own, without --overwrite.
func blocked(dest string, owned bool, opts Options) bool {
	if owned || opts.Overwrite {
		return false
	}
	return exists(dest)
}

func writeShim(projectDir string, o Output, dest string) (bool, error) {
	content, err := expectedShim(projectDir, o)
	if err != nil {
		return false, err
	}
	same := func() bool {
		prev, err := os.ReadFile(dest)
		return err == nil && bytes.Equal(prev, content)
	}
	return ensure(dest, same, func() error { return os.WriteFile(dest, content, 0o644) })
}

// ensure leaves dest alone when same() holds; otherwise it replaces dest
// via write. changed reports whether it wrote.
func ensure(dest string, same func() bool, write func() error) (changed bool, err error) {
	if same() {
		return false, nil
	}
	return true, replaceAt(dest, write)
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
	if o.Render == nil {
		content, ok, err := primitive.RenderForHost(projectDir, o.Prim)
		if err == nil && !ok {
			err = fmt.Errorf("no shim renderer for %s/%s", o.Kind, o.ID)
		}
		return content, err
	}
	data, err := os.ReadFile(filepath.Join(projectDir, o.Src))
	if err != nil {
		return nil, err
	}
	return o.Render(o.Prim, string(data))
}

// linkTarget points dest at the canonical target with a relative symlink.
func linkTarget(projectDir string, o Output, dest string) (bool, error) {
	rel, err := relLink(projectDir, o)
	if err != nil {
		return false, err
	}
	same := func() bool {
		cur, err := os.Readlink(dest)
		return err == nil && cur == rel
	}
	return ensure(dest, same, func() error { return os.Symlink(rel, dest) })
}

func relLink(projectDir string, o Output) (string, error) {
	return filepath.Rel(filepath.Dir(filepath.Join(projectDir, o.Dest)), filepath.Join(projectDir, o.Target))
}

// copyTarget mirrors the canonical file or directory at src into dest.
func copyTarget(src, dest string) (bool, error) {
	return ensure(dest, func() bool { return sameTree(src, dest) }, func() error { return copyTree(src, dest) })
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
		if filepath.Base(abs) == "SKILL.md" { // a per-skill dir keystone made
			_ = os.Remove(filepath.Dir(abs)) // only succeeds when empty
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
	case o.Strategy == Copy || mode == LinkCopy:
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
