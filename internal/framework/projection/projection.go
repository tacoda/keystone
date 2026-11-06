// Package projection plans and applies how the charter reaches each host
// (ADR 0010, docs/hosts.md). Every (host, kind) pair resolves to one
// strategy:
//
//   - native: the host reads the canonical file under .agents/ as-is.
//   - symlink: the host reads the same format at another path; keystone
//     links each entry (or copies it, in copy link mode).
//   - copy: as symlink, but always a copy (hosts whose symlink support is
//     unverified, or known to skip links).
//   - shim: the host needs a different format; keystone renders it.
//   - index: no host mechanism; AGENTS.md lists the primitive instead.
//   - report: a constraint the host cannot enforce; surfaced, never dropped.
//
// Planning is pure (Plan); writing is separate (Apply), so `keystone
// verify` can check drift (Check) against the same plan.
package projection

import (
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
	Copy    Strategy = "copy"
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
	// Target returns what a Symlink or Copy mirrors (repo-relative); nil
	// means the primitive's own file. A skill mirrors its whole directory.
	Target func(p primitive.Primitive) string
	// Render produces a Shim's bytes; nil means the Claude Code format
	// (primitive.RenderForHost).
	Render Renderer
}

// Renderer turns a primitive and its file contents into host bytes.
type Renderer func(p primitive.Primitive, fileContents string) ([]byte, error)

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
	Target   string // Symlink/Copy: canonical path the dest mirrors, repo-relative
	Strategy Strategy
	Prim     primitive.Primitive // the source primitive, for shim rendering
	Render   Renderer            // Shim renderer; nil = Claude Code format
}

// Warning flags a plan-level concern, e.g. a host that will see the same
// primitive at two paths (whether it dedupes by name is unverified).
type Warning struct {
	Host    string
	Message string
}

// PlanResult is a projection plan: the outputs, plus warnings about hosts
// that will see one primitive twice.
type PlanResult struct {
	Outputs  []Output
	Warnings []Warning
}

// Plan computes every output for the given hosts, in host order. Profiles
// earlier in the slice are planned first, so a later host that reads an
// earlier host's directory (Cursor and Copilot read .claude/) reuses it.
func Plan(primitives []primitive.Primitive, hosts []Profile) PlanResult {
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
	return PlanResult{Outputs: outs, Warnings: duplicateWarnings(primitives, hosts, located)}
}

// planOne plans p for host h. ok is false when nothing is needed: the
// host has no rule for the kind, the rule yields no dest, or the host
// already reads p at an existing location.
func planOne(h Profile, p primitive.Primitive, at []string) (Output, bool) {
	r, ok := h.Rules[primitive.Kind(p.Kind)]
	if !ok || r.Strategy == Native {
		return Output{}, false
	}
	o := Output{Host: h.Host, Kind: p.Kind, ID: p.ID, Src: p.Path, Strategy: r.Strategy, Prim: p, Render: r.Render}
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

// duplicateWarnings flags hosts that will load primitives from two
// directories they scan, one warning per host and kind. Whether a host
// dedupes by name is unverified (docs/hosts.md), so this is surfaced.
func duplicateWarnings(primitives []primitive.Primitive, hosts []Profile, located map[string][]string) []Warning {
	var ws []Warning
	for _, h := range hosts {
		counts := map[string]int{}
		for _, p := range primitives {
			if countRead(h.Reads[primitive.Kind(p.Kind)], located[key(p)]) > 1 {
				counts[p.Kind]++
			}
		}
		for _, k := range sortedKeys(counts) {
			ws = append(ws, Warning{Host: h.Host, Message: strconv.Itoa(counts[k]) + " " + k +
				"(s) visible at two paths this host reads (" + strings.Join(h.Reads[primitive.Kind(k)], ", ") +
				") — unverified whether it dedupes by name"})
		}
	}
	return ws
}

func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
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
