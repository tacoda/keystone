package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/tacoda/keystone/internal/framework/budget"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// runBudget handles `keystone budget [--json] [--dir D]`: the context an
// agent loads, per host (always loaded) and on demand, against the caps in
// keystone.json `budgets`.
func runBudget(args []string) error {
	dir, asJSON := ".", false
	help, err := parseFlags(args, map[string]*string{"--dir": &dir}, map[string]*bool{"--json": &asJSON},
		func() { printBudgetUsage(os.Stdout) })
	if help || err != nil {
		return err
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	r, err := measureBudget(absDir)
	if err != nil {
		return err
	}
	if asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(r)
	}
	printBudget(os.Stdout, r)
	return budgetError(r)
}

// measureBudget walks the charter and measures it for the enabled hosts.
func measureBudget(absDir string) (budget.Report, error) { return budget.MeasureProject(absDir) }

// budgetError fails when any cap is exceeded.
func budgetError(r budget.Report) error {
	if len(r.Violations) == 0 {
		return nil
	}
	return fmt.Errorf("%d context budget cap(s) exceeded", len(r.Violations))
}

func printBudget(w io.Writer, r budget.Report) {
	fmt.Fprintln(w, "Context budget (word-count estimate; real tokenizers count about 1.35× more)")
	fmt.Fprintln(w, "\nAlways loaded, per host:")
	for _, h := range r.Hosts {
		fmt.Fprintf(w, "  %-15s %6d\n", h.Host, h.Always)
		printEntries(w, h.Entries, 3, "      ")
	}
	total := 0
	for _, e := range r.OnDemand {
		total += e.Tokens
	}
	fmt.Fprintf(w, "\nOn demand (opened when they apply): %d\n", total)
	for _, k := range primitive.KnownKinds {
		if t := r.ByKind[string(k)]; t > 0 {
			fmt.Fprintf(w, "  %-15s %6d\n", k, t)
		}
	}
	fmt.Fprintln(w, "  largest:")
	printEntries(w, r.OnDemand, 5, "    ")
	printViolations(w, r.Violations)
}

func printEntries(w io.Writer, es []budget.Entry, n int, indent string) {
	for i, e := range es {
		if i == n {
			break
		}
		fmt.Fprintf(w, "%s%6d  %s (%s)\n", indent, e.Tokens, e.Path, e.Kind)
	}
}

// printViolations names each exceeded cap and the way to bring it down.
func printViolations(w io.Writer, vs []budget.Violation) {
	if len(vs) == 0 {
		return
	}
	fmt.Fprintf(w, "\n✗ %d cap(s) exceeded:\n", len(vs))
	for _, v := range vs {
		where := v.Scope
		if v.Path != "" {
			where += " " + v.Path
		}
		fmt.Fprintf(w, "  • %s: %d > %d (+%d)\n", where, v.Tokens, v.Cap, v.OverBy())
	}
	fmt.Fprintln(w, "  Run `keystone allocate` for root-file sections that can load on demand.")
}

// printBudgetSummary is the short form for init and doctor: one line per
// host, then any exceeded cap.
func printBudgetSummary(w io.Writer, r budget.Report) {
	fmt.Fprintln(w, "  context loaded every session (word-count estimate):")
	for _, h := range r.Hosts {
		fmt.Fprintf(w, "    %-15s %6d\n", h.Host, h.Always)
	}
	printViolations(w, r.Violations)
	fmt.Fprintln(w, "  (`keystone budget` for the breakdown)")
}

func printBudgetUsage(w io.Writer) {
	fmt.Fprint(w, `keystone budget — the context an agent loads

Usage:
    keystone budget [--json] [--dir <path>]

Always loaded, per host: root files (AGENTS.md, CLAUDE.md), rules without
paths: the host applies everywhere, and the name + description of every
skill, command, and subagent the host lists. On demand: rule, reference,
skill, command, and subagent bodies, opened when they apply.

Caps come from keystone.json "budgets":
    "budgets": {
      "always_loaded": {"max_tokens": 4000},
      "reference":     {"max_tokens": 50000, "max_tokens_per_load": 3000}
    }
A kind key caps that kind's on-demand total (max_tokens) and any one
file (max_tokens_per_load); 4.x keys (guides, corpus, …) still count.
Exits non-zero when a cap is exceeded; keystone verify checks the same.
`)
}
