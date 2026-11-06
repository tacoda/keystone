// Package projection plans and applies how the charter reaches each host
// (ADR 0010, docs/hosts.md). Every (host, kind) pair resolves to one
// strategy:
//
//   - native: the host reads the canonical file under .agents/ as-is.
//   - symlink: the host reads the same format at another path; keystone
//     links each entry (or copies it, in copy link mode).
//   - shim: the host needs a different format; keystone renders it.
//   - index: no host mechanism; AGENTS.md lists the primitive instead.
//   - report: a constraint the host cannot enforce; surfaced, never dropped.
//
// Planning is pure (Plan); writing is separate (Apply), so `keystone
// verify` can check drift (Check) against the same plan.
package projection

import (
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// Strategy is how one host reaches one primitive.
type Strategy string

const (
	Native  Strategy = "native"
	Symlink Strategy = "symlink"
	Shim    Strategy = "shim"
	Index   Strategy = "index"
	Report  Strategy = "report"
)

// LinkMode picks how a Symlink strategy is materialized.
type LinkMode string

const (
	LinkSymlink LinkMode = "symlink"
	LinkCopy    LinkMode = "copy"
)

// ResolveLinkMode maps the keystone.json `link_mode` value to a LinkMode.
// Empty means symlink, except on Windows, where symlinks need Developer
// Mode, so copy is the safer default.
func ResolveLinkMode(configured string) LinkMode {
	switch LinkMode(configured) {
	case LinkSymlink, LinkCopy:
		return LinkMode(configured)
	}
	if runtime.GOOS == "windows" {
		return LinkCopy
	}
	return LinkSymlink
}

// Rule is one host's handling of one kind.
type Rule struct {
	Strategy Strategy
	// Dest returns the repo-relative host path for p, or "" when p has no
	// projection on this host (e.g. a rule without paths:). Symlink and
	// Shim rules need it.
	Dest func(p primitive.Primitive) string
	// Target returns what a Symlink links to (repo-relative); nil means
	// the primitive's own file. A skill links its whole directory.
	Target func(p primitive.Primitive) string
}

// Profile is one host's projection table.
type Profile struct {
	Host  string
	Rules map[primitive.Kind]Rule
	// Reads lists, per kind, the repo-relative directories the host scans
	// and accepts as-is. The planner skips an output when the primitive
	// already sits under one of them (canonically, or written for an
	// earlier host).
	Reads map[primitive.Kind][]string
}

// Output is one planned host artifact.
type Output struct {
	Host     string
	Kind     string
	ID       string
	Src      string // canonical file, repo-relative
	Dest     string // host path, repo-relative ("" for index/report)
	Target   string // Symlink: canonical path the dest mirrors, repo-relative
	Strategy Strategy
	Prim     primitive.Primitive // the source primitive, for shim rendering
}

// Warning flags a plan-level concern, e.g. a host that will see the same
// primitive at two paths (whether it dedupes by name is unverified).
type Warning struct {
	Host    string
	Message string
}

// Plan computes every output for the given hosts, in host order. Profiles
// earlier in the slice are planned first, so a later host that reads an
// earlier host's directory (Cursor and Copilot read .claude/) reuses it.
func Plan(primitives []primitive.Primitive, hosts []Profile) ([]Output, []Warning) {
	var outs []Output
	located := map[string][]string{} // primitive key → repo-relative locations
	for _, p := range primitives {
		located[key(p)] = []string{p.Path}
	}
	for _, h := range hosts {
		for _, p := range primitives {
			o, ok := planOne(h, p, located[key(p)])
			if !ok {
				continue
			}
			outs = append(outs, o)
			if o.Dest != "" {
				located[key(p)] = append(located[key(p)], o.Dest)
			}
		}
	}
	return outs, duplicateWarnings(primitives, hosts, located)
}

// planOne plans p for host h. ok is false when nothing is needed: the
// host has no rule for the kind, the rule yields no dest, or the host
// already reads p at an existing location.
func planOne(h Profile, p primitive.Primitive, at []string) (Output, bool) {
	r, ok := h.Rules[primitive.Kind(p.Kind)]
	if !ok || r.Strategy == Native {
		return Output{}, false
	}
	o := Output{Host: h.Host, Kind: p.Kind, ID: p.ID, Src: p.Path, Strategy: r.Strategy, Prim: p}
	if r.Strategy == Index || r.Strategy == Report {
		return o, true
	}
	if readsAny(h.Reads[primitive.Kind(p.Kind)], at) {
		return Output{}, false
	}
	o.Dest = r.Dest(p)
	o.Target = p.Path
	if r.Target != nil {
		o.Target = r.Target(p)
	}
	return o, o.Dest != ""
}

// duplicateWarnings flags hosts that will load one primitive from two
// directories they scan.
func duplicateWarnings(primitives []primitive.Primitive, hosts []Profile, located map[string][]string) []Warning {
	var ws []Warning
	for _, h := range hosts {
		for _, p := range primitives {
			if n := countRead(h.Reads[primitive.Kind(p.Kind)], located[key(p)]); n > 1 {
				ws = append(ws, Warning{Host: h.Host, Message: p.Kind + "/" + p.ID + " is visible at " + strconv.Itoa(n) + " paths this host reads"})
			}
		}
	}
	return ws
}

func readsAny(dirs, at []string) bool { return countRead(dirs, at) > 0 }

func countRead(dirs, at []string) int {
	n := 0
	for _, loc := range at {
		for _, d := range dirs {
			if underDir(loc, d) {
				n++
				break
			}
		}
	}
	return n
}

// underDir reports whether repo-relative loc sits under dir.
func underDir(loc, dir string) bool {
	return strings.HasPrefix(filepath.ToSlash(loc), strings.TrimSuffix(dir, "/")+"/")
}

func key(p primitive.Primitive) string { return p.Kind + "/" + p.ID }

// skillDir returns the canonical directory of a skill (the parent of
// SKILL.md), repo-relative.
func skillDir(p primitive.Primitive) string { return path.Dir(filepath.ToSlash(p.Path)) }

// Dests returns the sorted, de-duplicated host paths of outs.
func Dests(outs []Output) []string {
	seen := map[string]bool{}
	var ds []string
	for _, o := range outs {
		if o.Dest != "" && !seen[o.Dest] {
			seen[o.Dest] = true
			ds = append(ds, o.Dest)
		}
	}
	sort.Strings(ds)
	return ds
}
