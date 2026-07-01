package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tacoda/keystone/internal/framework/charter"
	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// runCharter dispatches `keystone charter <sub>`: `coverage` (files no
// guide governs) and `show` (the roster, optionally post-cascade).
func runCharter(args []string) error {
	if len(args) == 0 {
		printCharterUsage(os.Stderr)
		return fmt.Errorf("`keystone charter` requires a subcommand (coverage | show)")
	}
	switch args[0] {
	case "help", "--help", "-h":
		printCharterUsage(os.Stdout)
		return nil
	case "coverage":
		return runCharterCoverage(args[1:])
	case "show":
		return runCharterShow(args[1:])
	case "conformance":
		return runCharterConformance(args[1:])
	default:
		return fmt.Errorf("unknown charter subcommand %q (use: coverage | show | conformance)", args[0])
	}
}

func runCharterConformance(args []string) error {
	absDir, err := filepath.Abs(dirArg(args))
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	rub, err := charter.Conformance(absDir, config.DefaultCharterRoot)
	if err != nil {
		return fmt.Errorf("conformance: %w", err)
	}
	fmt.Fprintf(os.Stdout, "Charter conformance: %s\n\n", rub.Verdict)
	for _, c := range rub.Criteria {
		fmt.Fprintf(os.Stdout, "  %-8s %-22s %s\n", c.Status, c.Name, c.Detail)
	}
	return nil
}

func runCharterCoverage(args []string) error {
	absDir, err := filepath.Abs(dirArg(args))
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	res, err := charter.Coverage(absDir, config.DefaultCharterRoot)
	if err != nil {
		return fmt.Errorf("coverage: %w", err)
	}
	printCoverage(res)
	return nil
}

func printCoverage(r charter.CoverageResult) {
	pct := 100
	if r.Total > 0 {
		pct = r.Governed * 100 / r.Total
	}
	fmt.Fprintf(os.Stdout, "Charter coverage — %d files, %d governed, %d uncharted (%d%% governed)\n",
		r.Total, r.Governed, len(r.Uncharted), pct)
	if len(r.Uncharted) == 0 {
		fmt.Fprintln(os.Stdout, "\nEvery scanned file is governed by a guide.")
		return
	}
	counts := r.UnchartedByRegion()
	fmt.Fprintln(os.Stdout, "\nUncharted regions (no guide globs match):")
	for _, region := range sortByCountDesc(counts) {
		fmt.Fprintf(os.Stdout, "  %-28s %d\n", region, counts[region])
	}
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

// showOpts is the parsed `charter show` invocation.
type showOpts struct {
	dir       string
	kind      string
	effective bool
}

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
	kept := filterByKind(prims, opts.kind)
	entries := allEntries(kept)
	if opts.effective {
		entries = charter.Effective(kept)
	}
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

func allEntries(prims []primitive.Primitive) []charter.Entry {
	out := make([]charter.Entry, 0, len(prims))
	for _, p := range prims {
		out = append(out, charter.Entry{Primitive: p})
	}
	return out
}

func printRoster(entries []charter.Entry, effective bool) {
	label := "charter roster"
	if effective {
		label = "effective charter (post-cascade)"
	}
	fmt.Fprintf(os.Stdout, "%s — %d primitive(s)\n", label, len(entries))
	byKind := map[string][]charter.Entry{}
	for _, e := range entries {
		byKind[e.Primitive.Kind] = append(byKind[e.Primitive.Kind], e)
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

func printRosterLine(e charter.Entry) {
	src := ""
	if e.Primitive.Provenance != "" && e.Primitive.Provenance != "project" {
		src = "  [" + e.Primitive.Provenance + "]"
	}
	shadow := ""
	if len(e.Shadows) > 0 {
		shadow = "  (overrides " + strings.Join(e.Shadows, ", ") + ")"
	}
	fmt.Fprintf(os.Stdout, "  %-40s %s%s%s\n", e.Primitive.ID, e.Primitive.Description, src, shadow)
}

func printCharterUsage(w *os.File) {
	fmt.Fprint(w, `keystone charter — inspect the charter

Usage:

    keystone charter coverage [--dir D]                        # files no guide governs (uncharted territory)
    keystone charter show [--effective] [--kind K] [--dir D]   # the charter roster
    keystone charter conformance [--dir D]                     # rubric: does the repo conform to its charter?

Coverage reports which project files a guide's globs match and which are
uncharted — where the agent runs with no ambient rule. Show lists the
charter's primitives by kind; --effective resolves the cascade (project
wins, then policies) and annotates what each winner overrides.
`)
}
