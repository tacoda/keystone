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
	default:
		return fmt.Errorf("unknown charter subcommand %q (use: coverage)", args[0])
	}
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
	total, governed, uncharted, err := scanCoverage(absDir, globs)
	if err != nil {
		return err
	}
	printCoverage(total, governed, uncharted)
	return nil
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
func scanCoverage(absDir string, globs []string) (total, governed int, uncharted []string, err error) {
	err = filepath.WalkDir(absDir, func(path string, d fs.DirEntry, walkErr error) error {
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
		total++
		if anyGlobMatches(globs, rel) {
			governed++
		} else {
			uncharted = append(uncharted, rel)
		}
		return nil
	})
	return total, governed, uncharted, err
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
func printCoverage(total, governed int, uncharted []string) {
	pct := 100
	if total > 0 {
		pct = governed * 100 / total
	}
	fmt.Fprintf(os.Stdout, "Charter coverage — %d files, %d governed, %d uncharted (%d%% governed)\n",
		total, governed, len(uncharted), pct)
	if len(uncharted) == 0 {
		fmt.Fprintln(os.Stdout, "\nEvery scanned file is governed by a guide.")
		return
	}
	counts := map[string]int{}
	for _, rel := range uncharted {
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

    keystone charter coverage [--dir D]   # files no guide governs (uncharted territory)

Coverage walks the project's source files and reports which are matched
by a guide's globs and which are uncharted — where the agent runs with
no ambient rule. Generated/vendored/hidden trees are skipped.
`)
}
