package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tacoda/keystone/internal/framework/charter"
	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// runCharter dispatches `keystone charter <sub>`: `query` (what applies
// here?), `coverage` (files no rule governs), `show` (the roster,
// optionally post-cascade), and `conformance`.
func runCharter(args []string) error {
	if len(args) == 0 {
		printCharterUsage(os.Stderr)
		return fmt.Errorf("`keystone charter` requires a subcommand (query | coverage | show | conformance)")
	}
	switch args[0] {
	case "help", "--help", "-h":
		printCharterUsage(os.Stdout)
		return nil
	case "query":
		return runCharterQuery(args[1:])
	case "move":
		return runCharterMove(args[1:])
	case "promote":
		return runCharterPromote(args[1:])
	case "demote":
		return runCharterDemote(args[1:])
	case "coverage":
		return runCharterCoverage(args[1:])
	case "show":
		return runCharterShow(args[1:])
	case "conformance":
		return runCharterConformance(args[1:])
	default:
		return fmt.Errorf("unknown charter subcommand %q (use: query | coverage | show | conformance)", args[0])
	}
}

// runCharterQuery handles `keystone charter query [files…] [--task <text>]
// [--full]`. Composes "what context applies here?" through
// charter.Query — the same helper the MCP server's keystone_charter_query
// calls, so CLI and MCP never disagree.
func runCharterQuery(args []string) error {
	var files []string
	task, asJSON, full := "", false, false
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--task":
			i++
			if i >= len(args) {
				return fmt.Errorf("--task requires a value")
			}
			task = args[i]
		case a == "--json":
			asJSON = true
		case a == "--full":
			full = true
		case a == "--dir":
			i++ // value consumed here; dirArg re-reads it below
		case strings.HasPrefix(a, "--dir="), strings.HasPrefix(a, "--task="):
			// --dir handled by dirArg; --task by the --task sweep
		default:
			files = append(files, a)
		}
	}
	if task == "" {
		for i := 0; i+1 < len(args); i++ {
			if args[i] == "--task" || strings.HasPrefix(args[i], "--task=") {
				if strings.HasPrefix(args[i], "--task=") {
					task = strings.TrimPrefix(args[i], "--task=")
				} else {
					task = args[i+1]
				}
			}
		}
	}
	absDir, err := filepath.Abs(dirArg(args))
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	prims, _, err := primitive.Walk(absDir, config.DefaultAgentsRoot)
	if err != nil {
		return err
	}
	composed, _ := primitive.Compose(prims)
	res := charter.Query(composed, files, task)

	if asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		return enc.Encode(res.Entries)
	}
	if len(res.Entries) == 0 {
		fmt.Fprintln(os.Stdout, "(no context matches)")
		return nil
	}
	for _, e := range res.Entries {
		fmt.Fprintf(os.Stdout, "%s/%s  %s\n", e.Kind, e.ID, e.Description)
		fmt.Fprintf(os.Stdout, "  %s  [%s]\n", e.Path, e.Provenance)
		if e.Why != "everywhere" {
			fmt.Fprintf(os.Stdout, "  matched: %s\n", e.Why)
		}
		if full {
			body, rerr := os.ReadFile(filepath.Join(absDir, filepath.FromSlash(e.Path)))
			if rerr != nil {
				fmt.Fprintf(os.Stderr, "  ! read %s: %v\n", e.Path, rerr)
				continue
			}
			fmt.Fprintf(os.Stdout, "\n%s\n\n", strings.TrimRight(string(body), "\n"))
		}
	}
	return nil
}

func runCharterConformance(args []string) error {
	absDir, err := filepath.Abs(dirArg(args))
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	rub, err := charter.Conformance(absDir, config.DefaultAgentsRoot)
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
	res, err := charter.Coverage(absDir, config.DefaultAgentsRoot)
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
		fmt.Fprintln(os.Stdout, "\nEvery scanned file is governed by a rule.")
		return
	}
	counts := r.UnchartedByRegion()
	fmt.Fprintln(os.Stdout, "\nUncharted regions (no rule paths: match):")
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
	prims, _, err := primitive.Walk(absDir, config.DefaultAgentsRoot)
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
		if primitive.IsKind(p.Kind, kind) {
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
	sort.SliceStable(entries, func(i, j int) bool {
		return primitive.KindRank(entries[i].Primitive.Kind) < primitive.KindRank(entries[j].Primitive.Kind)
	})
	kind := ""
	for _, e := range entries {
		if e.Primitive.Kind != kind {
			kind = e.Primitive.Kind
			fmt.Fprintf(os.Stdout, "\n# %s\n", kind)
		}
		printRosterLine(e)
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
	fmt.Fprint(w, `keystone charter — query, move, inspect

Usage:

    keystone charter query [files…] [--task TEXT] [--full] [--json]  # what context applies here?
    keystone charter move <kind>/<id> <new-kind>[/<new-path>]        # relocate or re-kind a primitive
    keystone charter promote <kind>/<id>                             # corpus → reference → rule, one step up
    keystone charter demote <kind>/<id>                              # rule → reference → corpus, one step down
    keystone charter coverage [--dir D] [--full] [--dir D]           # files no rule governs (uncharted territory)
    keystone charter show [--effective] [--kind K] [--dir D]         # the charter roster
    keystone charter conformance [--dir D] [--full] [--dir D]        # rubric: does the repo conform to its charter?

Query composes "what applies here?": rules without paths: (everywhere),
path-matched rules and corpus, cited references, then task-scored
entries -- by keyword, descriptors only until --full. Move relocates a
primitive and rewrites citations that point at it; promote/demote walk
the information ladder (the what → the why → the how). Capability kinds
move laterally -- there is no ladder between them.
`)
}
