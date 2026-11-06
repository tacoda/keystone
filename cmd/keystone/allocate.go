package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/tacoda/keystone/internal/framework/allocate"
	"github.com/tacoda/keystone/internal/framework/lockfile"
)

// runAllocate handles `keystone allocate [--apply] [--dir D]`: report, and
// with --apply write, where each piece of an existing host-specific setup
// goes in .agents/. Originals are never changed.
func runAllocate(args []string) error {
	dir, apply := ".", false
	help, err := parseFlags(args, map[string]*string{"--dir": &dir}, map[string]*bool{"--apply": &apply},
		func() { printAllocateUsage(os.Stdout) })
	if help || err != nil {
		return err
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	lf, err := lockfile.Read(absDir)
	if err != nil {
		return err
	}
	items, err := allocate.Plan(absDir, lf.Projections, lf.ManagedEntries)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		fmt.Fprintln(os.Stdout, "✓ keystone allocate — nothing to move: no host-specific setup outside keystone's own projections")
		return nil
	}
	printAllocation(os.Stdout, items)
	if !apply {
		fmt.Fprintln(os.Stdout, "\nRun `keystone allocate --apply` to write these into .agents/ (originals stay in place).")
		return nil
	}
	res, err := allocate.Apply(absDir, items)
	printApplied(os.Stdout, res)
	return err
}

func printAllocation(w io.Writer, items []allocate.Item) {
	moves := 0
	for _, it := range items {
		if it.Moves() {
			moves++
		}
	}
	fmt.Fprintf(w, "keystone allocate — %d item(s) move into .agents/, %d stay\n\n", moves, len(items)-moves)
	for _, it := range items {
		printItem(w, it)
	}
}

func printItem(w io.Writer, it allocate.Item) {
	if !it.Moves() {
		fmt.Fprintf(w, "  ! %s — %s\n", it.Source, it.Note)
		return
	}
	fmt.Fprintf(w, "  %s → %s %s\n", it.Source, it.Kind, it.Target)
	if it.Note != "" {
		fmt.Fprintf(w, "      %s\n", it.Note)
	}
}

func printApplied(w io.Writer, res allocate.Result) {
	for _, p := range res.Wrote {
		fmt.Fprintf(w, "  wrote: %s\n", p)
	}
	for _, p := range res.Exists {
		fmt.Fprintf(w, "  exists, kept: %s\n", p)
	}
	fmt.Fprintf(w, "✓ keystone allocate — %d written, %d already present\n", len(res.Wrote), len(res.Exists))
	if len(res.Wrote) > 0 {
		fmt.Fprintln(w, "Next: `keystone lint`, then `keystone index && keystone project`. `project` refuses a host path that")
		fmt.Fprintln(w, "still holds an original; remove the original once you have checked the converted file.")
	}
}

func printAllocateUsage(w io.Writer) {
	fmt.Fprint(w, `keystone allocate — move an existing agent setup into .agents/

Usage:
    keystone allocate [--apply] [--dir <path>]

Reads every host-specific setup — .claude/ skills, agents, commands, rules,
settings.json permissions and hooks; .cursor/rules and .cursorrules;
Copilot instructions, prompts, and agents; .clinerules, .windsurfrules;
.continue/; .opencode/ — and reports, per item, the 5.0 kind and the
.agents/ path it belongs at. Keystone's own projections are skipped.

  (default)   report only; nothing is written
  --apply     write the converted files into .agents/. An existing .agents/
              file is never overwritten, and no original is changed.
`)
}
