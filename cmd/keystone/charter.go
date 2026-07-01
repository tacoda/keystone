package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// runCharter dispatches `keystone charter <sub>`. Today: `coverage` —
// which project files no guide governs ("uncharted territory").
func runCharter(args []string) error {
	if len(args) == 0 {
		printCharterUsage(os.Stderr)
		return fmt.Errorf("`keystone charter` requires a subcommand (coverage)")
	}
	switch args[0] {
	case "help", "--help", "-h":
		printCharterUsage(os.Stdout)
		return nil
	case "coverage":
		return runCharterCoverage(args[1:])
	case "show":
		return runCharterShow(args[1:])
	default:
		return fmt.Errorf("unknown charter subcommand %q (use: coverage | show)", args[0])
	}
}

// showOpts is the parsed `charter show` invocation.
type showOpts struct {
	dir       string
	kind      string
	effective bool
}

// runCharterShow renders the charter roster. With --effective it
// resolves the cascade — one winning primitive per id (project wins,
// then policies in order) — annotating anything it shadows. Without it,
// every layer's primitive is listed.
func runCharterShow(args []string) error {
	opts := parseShowOpts(args)
	absDir, err := filepath.Abs(opts.dir)
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	prims, _, err := primitive.Walk(absDir, config.DefaultCharterRoot)
	if err != nil {
		return fmt.Errorf("walk charter: %w", err)
	}
	entries := rosterEntries(prims, opts)
	printRoster(entries, opts.effective)
	return nil
}

func parseShowOpts(args []string) showOpts {
	opts := showOpts{dir: "."}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--effective":
			opts.effective = true
		case "--dir":
			if i+1 < len(args) {
				opts.dir = args[i+1]
				i++
			}
		case "--kind":
			if i+1 < len(args) {
				opts.kind = args[i+1]
				i++
			}
		}
	}
	return opts
}

// rosterEntry is one line of the roster: the winning primitive plus the
// layers it shadows (empty unless --effective resolved an override).
type rosterEntry struct {
	p       primitive.Primitive
	shadows []string
}

// rosterEntries filters by --kind and, when effective, collapses each id
// to its cascade winner (project over policy) recording what it shadows.
func rosterEntries(prims []primitive.Primitive, opts showOpts) []rosterEntry {
	kept := filterByKind(prims, opts.kind)
	if !opts.effective {
		out := make([]rosterEntry, 0, len(kept))
		for _, p := range kept {
			out = append(out, rosterEntry{p: p})
		}
		return out
	}
	return resolveEffective(kept)
}

func filterByKind(prims []primitive.Primitive, kind string) []primitive.Primitive {
	if kind == "" {
		return prims
	}
	var out []primitive.Primitive
	for _, p := range prims {
		if p.Kind == kind {
			out = append(out, p)
		}
	}
	return out
}

// resolveEffective collapses same-id primitives to their cascade winner
// (project over policy), preserving first-seen order.
func resolveEffective(prims []primitive.Primitive) []rosterEntry {
	byID := map[string]*rosterEntry{}
	var order []string
	for _, p := range prims {
		key := p.Kind + "/" + p.ID
		if e, ok := byID[key]; ok {
			mergeCascade(e, p)
			continue
		}
		byID[key] = &rosterEntry{p: p}
		order = append(order, key)
	}
	out := make([]rosterEntry, 0, len(order))
	for _, k := range order {
		out = append(out, *byID[k])
	}
	return out
}

// mergeCascade resolves a same-id conflict into e: the project layer
// wins over any policy; the loser is recorded as shadowed.
func mergeCascade(e *rosterEntry, p primitive.Primitive) {
	if e.p.Provenance != "project" && p.Provenance == "project" {
		e.shadows = append(e.shadows, e.p.Provenance)
		e.p = p
		return
	}
	e.shadows = append(e.shadows, p.Provenance)
}

// printRoster prints entries grouped by kind, sorted, with provenance.
func printRoster(entries []rosterEntry, effective bool) {
	label := "charter roster"
	if effective {
		label = "effective charter (post-cascade)"
	}
	fmt.Fprintf(os.Stdout, "%s — %d primitive(s)\n", label, len(entries))
	byKind := map[string][]rosterEntry{}
	for _, e := range entries {
		byKind[e.p.Kind] = append(byKind[e.p.Kind], e)
	}
	kinds := make([]string, 0, len(byKind))
	for k := range byKind {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)
	for _, k := range kinds {
		fmt.Fprintf(os.Stdout, "\n%s\n", k)
		for _, e := range byKind[k] {
			printRosterLine(e)
		}
	}
}

func printRosterLine(e rosterEntry) {
	src := ""
	if e.p.Provenance != "" && e.p.Provenance != "project" {
		src = "  [" + e.p.Provenance + "]"
	}
	shadow := ""
	if len(e.shadows) > 0 {
		shadow = "  (overrides " + strings.Join(e.shadows, ", ") + ")"
	}
	fmt.Fprintf(os.Stdout, "  %-40s %s%s%s\n", e.p.ID, e.p.Description, src, shadow)
}

// runCharterCoverage reports which files in the project a guide governs
// and which are uncharted — matched by no guide's globs. A charter has
// jurisdiction; this surfaces where it's silent, so the agent runs
// unconstrained there.
func runCharterCoverage(args []string) error {
	absDir, err := filepath.Abs(dirArg(args))
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	globs, err := collectGuideGlobs(absDir)
	if err != nil {
		return err
	}
	res, err := scanCoverage(absDir, globs)
	if err != nil {
		return err
	}
	printCoverage(res)
	return nil
}

// coverageResult is the outcome of a coverage scan.
type coverageResult struct {
	total, governed int
	uncharted       []string
}

// collectGuideGlobs returns every glob any guide claims, after cascade
// composition.
func collectGuideGlobs(absDir string) ([]string, error) {
	prims, _, err := primitive.Walk(absDir, config.DefaultCharterRoot)
	if err != nil {
		return nil, fmt.Errorf("walk charter: %w", err)
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

// scanCoverage walks project source files, classifying each as governed
// (matched by ≥1 guide glob) or uncharted. Generated/vendored/hidden
// trees are skipped — coverage is about the source an agent edits.
func scanCoverage(absDir string, globs []string) (coverageResult, error) {
	var r coverageResult
	err := filepath.WalkDir(absDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil // skip unreadable entries, don't abort
		}
		if d.IsDir() {
			if path != absDir && skipCoverageDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		rel, e := filepath.Rel(absDir, path)
		if e != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		r.total++
		if anyGlobMatches(globs, rel) {
			r.governed++
		} else {
			r.uncharted = append(r.uncharted, rel)
		}
		return nil
	})
	return r, err
}

// skipCoverageDir reports whether a directory is out of scope for
// coverage: hidden dirs (.git, .charter, projected .claude/.cursor/…)
// and common generated/vendored trees.
func skipCoverageDir(name string) bool {
	if strings.HasPrefix(name, ".") {
		return true
	}
	switch name {
	case "node_modules", "vendor", "dist", "build":
		return true
	}
	return false
}

func anyGlobMatches(globs []string, rel string) bool {
	for _, g := range globs {
		if primitive.MatchPath(g, rel) {
			return true
		}
	}
	return false
}

// printCoverage renders the summary + uncharted regions grouped by their
// top-level directory (most-uncharted first).
func printCoverage(r coverageResult) {
	pct := 100
	if r.total > 0 {
		pct = r.governed * 100 / r.total
	}
	fmt.Fprintf(os.Stdout, "Charter coverage — %d files, %d governed, %d uncharted (%d%% governed)\n",
		r.total, r.governed, len(r.uncharted), pct)
	if len(r.uncharted) == 0 {
		fmt.Fprintln(os.Stdout, "\nEvery scanned file is governed by a guide.")
		return
	}
	counts := map[string]int{}
	for _, rel := range r.uncharted {
		counts[topSegment(rel)]++
	}
	fmt.Fprintln(os.Stdout, "\nUncharted regions (no guide globs match):")
	for _, region := range sortByCountDesc(counts) {
		fmt.Fprintf(os.Stdout, "  %-28s %d\n", region, counts[region])
	}
}

// topSegment returns the first path segment (a directory), or "(root)"
// for a top-level file.
func topSegment(rel string) string {
	if i := strings.Index(rel, "/"); i >= 0 {
		return rel[:i] + "/"
	}
	return "(root)"
}

func sortByCountDesc(counts map[string]int) []string {
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if counts[keys[i]] != counts[keys[j]] {
			return counts[keys[i]] > counts[keys[j]]
		}
		return keys[i] < keys[j]
	})
	return keys
}

func printCharterUsage(w *os.File) {
	fmt.Fprint(w, `keystone charter — inspect the charter

Usage:

    keystone charter coverage [--dir D]              # files no guide governs (uncharted territory)
    keystone charter show [--effective] [--kind K] [--dir D]   # the charter roster

Coverage walks the project's source files and reports which are matched
by a guide's globs and which are uncharted — where the agent runs with
no ambient rule. Generated/vendored/hidden trees are skipped.

Show lists the charter's primitives grouped by kind. With --effective it
resolves the cascade — one winning primitive per id (project wins, then
policies in order) — and annotates anything a winner overrides.
`)
}
