// Package drift implements context drift (spec pillar 2): the gap
// between what the charter says and what the code is. Three checks —
// dead links, paths: globs that match nothing, and stale verified: dates
// — reported as warnings. Keystone reports; agents and people judge
// (decisions 12/16: no verdicts, no auto-stamping, no auto-refresh).
package drift

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
	"github.com/tacoda/keystone/internal/framework/projection"
)

// Finding kinds.
const (
	KindDeadLink  = "dead-link"
	KindPathsGone = "paths-gone"
	KindStale     = "stale"
)

// Finding is one drift warning, named by file and reason.
type Finding struct {
	Kind   string `json:"kind"`      // dead-link | paths-gone | stale
	Path   string `json:"path"`      // repo-relative primitive file
	PrimID string `json:"primitive"` // kind/id of the drifting primitive
	Detail string `json:"detail"`    // link target, glob list, or verified date
	Reason string `json:"reason"`    // human-readable explanation
}

// Report is the outcome of a Scan.
type Report struct {
	Checked  int       `json:"checked"`  // primitives examined
	Warnings []Finding `json:"warnings"` // sorted by path, then kind
}

// Options scopes a Scan.
type Options struct {
	// Files, when non-empty, restricts the scan to primitives relevant
	// to those repo-relative files: a primitive is checked when one of
	// its paths: globs matches a listed file, or the primitive's own
	// file is listed.
	Files []string
	// Now is the reference time for staleness (tests inject it).
	Now time.Time
	// Days is the staleness window for verified: dates.
	Days int
}

// Scan walks the charter at projectDir and reports drift. Only
// project-owned primitives are checked — policy-vendored and adopted
// repo docs are not the project's to re-verify here.
func Scan(projectDir string, opts Options) (Report, error) {
	var rep Report
	if opts.Days <= 0 {
		opts.Days = config.DefaultDriftDays
	}
	if opts.Now.IsZero() {
		opts.Now = time.Now()
	}
	prims, _, err := primitive.Walk(projectDir, config.DefaultAgentsRoot)
	if err != nil {
		return rep, err
	}
	for _, p := range prims {
		if p.Provenance != "project" {
			continue
		}
		if !inScope(p, opts.Files) {
			continue
		}
		rep.Checked++
		fs, err := scanPrimitive(projectDir, p, opts)
		if err != nil {
			return rep, err
		}
		rep.Warnings = append(rep.Warnings, fs...)
	}
	sort.Slice(rep.Warnings, func(i, j int) bool {
		if rep.Warnings[i].Path != rep.Warnings[j].Path {
			return rep.Warnings[i].Path < rep.Warnings[j].Path
		}
		return rep.Warnings[i].Kind < rep.Warnings[j].Kind
	})
	return rep, nil
}

// inScope reports whether a primitive is relevant to the given files.
// An empty file list means everything is in scope.
func inScope(p primitive.Primitive, files []string) bool {
	if len(files) == 0 {
		return true
	}
	for _, f := range files {
		if filepath.ToSlash(f) == p.Path {
			return true
		}
		for _, g := range p.Paths {
			if primitive.MatchPath(g, filepath.ToSlash(f)) {
				return true
			}
		}
	}
	return false
}

// scanPrimitive runs the three checks against one primitive.
func scanPrimitive(projectDir string, p primitive.Primitive, opts Options) ([]Finding, error) {
	var out []Finding
	primRef := p.Kind + "/" + p.ID
	base := Finding{Path: p.Path, PrimID: primRef}

	// 1. Dead links: relative links in the body whose target is gone.
	abs := filepath.Join(projectDir, filepath.FromSlash(p.Path))
	data, err := os.ReadFile(abs)
	if err != nil {
		return nil, err
	}
	_, body, _ := primitive.SplitFrontmatter(string(data))
	for _, target := range projection.RelativeLinkTargets([]byte(body), p.Path) {
		if _, err := os.Stat(filepath.Join(projectDir, filepath.FromSlash(target))); err != nil {
			if os.IsNotExist(err) {
				f := base
				f.Kind, f.Detail = KindDeadLink, target
				f.Reason = fmt.Sprintf("relative link %q does not exist", target)
				out = append(out, f)
			}
		}
	}

	// 2. Named paths gone: every positive paths: glob matches zero files.
	if len(p.Paths) > 0 {
		total := 0
		var positive []string
		for _, g := range p.Paths {
			if strings.HasPrefix(g, "!") {
				continue
			}
			positive = append(positive, g)
			matches, err := primitive.GlobRepo(projectDir, g)
			if err != nil {
				return out, err
			}
			total += len(matches)
		}
		if total == 0 {
			f := base
			f.Kind, f.Detail = KindPathsGone, strings.Join(positive, ", ")
			f.Reason = "paths: match zero files — the governed region may have vanished"
			out = append(out, f)
		}
	}

	// 3. Staleness: a verified: date older than the window. Author-set
	// only; drift never stamps.
	if p.Verified != "" {
		v, err := time.Parse("2006-01-02", p.Verified)
		if err != nil {
			f := base
			f.Kind, f.Detail = KindStale, p.Verified
			f.Reason = "verified: is not a YYYY-MM-DD date"
			out = append(out, f)
		} else if opts.Now.Sub(v) > time.Duration(opts.Days)*24*time.Hour {
			f := base
			f.Kind, f.Detail = KindStale, p.Verified
			f.Reason = fmt.Sprintf("verified %s is older than the %d-day window", p.Verified, opts.Days)
			out = append(out, f)
		}
	}
	return out, nil
}
