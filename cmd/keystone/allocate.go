package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/tacoda/keystone/internal/framework/allocate"
	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/lockfile"
	"github.com/tacoda/keystone/internal/framework/projection"
)

// runAllocate handles `keystone allocate [--apply] [--dir D]`: report, and
// with --apply write, where each piece of an existing host-specific setup
// goes in .agents/. Originals are never changed.
func runAllocate(args []string) error {
	dir, apply, overwrite := ".", false, false
	help, err := parseFlags(args, map[string]*string{"--dir": &dir},
		map[string]*bool{"--apply": &apply, "--overwrite": &overwrite},
		func() { printAllocateUsage(os.Stdout) })
	if help || err != nil {
		return err
	}
	absDir, lf, items, err := planAllocation(dir)
	if err != nil {
		return err
	}
	if !reportAllocation(absDir, items, apply || overwrite) {
		return nil
	}
	res, err := allocate.Apply(absDir, items)
	if err != nil || !overwrite {
		printApplied(os.Stdout, res)
		return err
	}
	return takeOver(absDir, lf, items, res)
}

// reportAllocation prints the plan and reports whether there is anything
// to write (and writing was asked for).
func reportAllocation(absDir string, items []allocate.Item, write bool) bool {
	printRootTotals(os.Stdout, allocate.RootTotals(absDir, items))
	if len(items) == 0 {
		fmt.Fprintln(os.Stdout, "✓ keystone allocate — nothing to move: no host-specific setup outside keystone's projections, and no root-file section that loads better on demand")
		return false
	}
	printAllocation(os.Stdout, items)
	if !write {
		fmt.Fprintln(os.Stdout, "\nRun `keystone allocate --apply` to write these into .agents/ (originals stay in place).")
	}
	return write
}

// planAllocation resolves dir, reads the lockfile (keystone's own output
// is skipped), and plans every item.
func planAllocation(dir string) (string, *lockfile.Lockfile, []allocate.Item, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", nil, nil, fmt.Errorf("resolve dir: %w", err)
	}
	lf, err := lockfile.Read(absDir)
	if err != nil {
		return "", nil, nil, err
	}
	items, err := allocate.Plan(absDir, lf.Projections, lf.ManagedEntries)
	return absDir, lf, items, err
}

// takeOver is --overwrite: keystone takes ownership of the originals whose
// converted files were written, then re-indexes and re-projects so each
// host reads the charter instead.
func takeOver(absDir string, lf *lockfile.Lockfile, items []allocate.Item, res allocate.Result) error {
	own, err := allocate.Overwrite(absDir, items, res.Wrote, enabledHosts(absDir))
	if err != nil {
		return err
	}
	lf.Projections = append(lf.Projections, own.Owned...)
	recordManaged(lf, mergeManaged(lf.ManagedEntries, own.Managed))
	if err := lockfile.Write(absDir, lf); err != nil {
		return err
	}
	for _, p := range res.Exists {
		fmt.Fprintf(os.Stdout, "  exists, original kept: %s\n", p)
	}
	for _, k := range own.Kept {
		fmt.Fprintf(os.Stdout, "  original kept: %s\n", k)
	}
	fmt.Fprintf(os.Stdout, "✓ keystone allocate --overwrite — %d written; %d original(s) now keystone's\n", len(res.Wrote), len(own.Owned))
	return indexAndProject(absDir, config.DefaultAgentsRoot, false)
}

// enabledHosts is the hosts keystone projects to (keystone.json adapters,
// plus claude-code, which is always on).
func enabledHosts(absDir string) []string {
	var adapters []string
	if cfg, err := config.ReadProjectConfig(absDir); err == nil && cfg != nil {
		adapters = cfg.Adapters
	}
	var hosts []string
	for _, p := range projection.Enabled(adapters) {
		hosts = append(hosts, p.Host)
	}
	return hosts
}

// mergeManaged appends add's entries to the lockfile's, per file.
func mergeManaged(have, add map[string][]string) map[string][]string {
	out := map[string][]string{}
	for dest, entries := range add {
		out[dest] = append(append([]string(nil), have[dest]...), entries...)
	}
	return out
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

func printRootTotals(w io.Writer, totals []allocate.RootTotal) {
	for _, t := range totals {
		if t.Tokens == 0 {
			continue
		}
		fmt.Fprintf(w, "  %s: ~%d tokens of your content loaded every session", t.File, t.Tokens)
		if t.Saved > 0 {
			fmt.Fprintf(w, "; the moves below save ~%d", t.Saved)
		}
		fmt.Fprintln(w)
	}
}

func printItem(w io.Writer, it allocate.Item) {
	if !it.Moves() {
		fmt.Fprintf(w, "  ! %s — %s\n", it.Source, it.Note)
		return
	}
	fmt.Fprintf(w, "  %s → %s %s\n", it.Source, it.Kind, it.Target)
	if it.Tokens > 0 {
		fmt.Fprintf(w, "      saves ~%d always-loaded tokens\n", it.Tokens)
	}
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

Reads the root files (CLAUDE.md, AGENTS.md, GEMINI.md, .goosehints) and
proposes moving each section that need not load every session: a section
naming file patterns becomes a path-scoped rule, a numbered procedure a
command, a "before you commit, run X" line a hook, long reasoning a
reference. Iron laws, security, and keystone's managed block stay.

Also reads every host-specific setup — .claude/ skills, agents, commands, rules,
settings.json permissions and hooks; .cursor/rules and .cursorrules;
Copilot instructions, prompts, and agents; .clinerules, .windsurfrules;
.continue/; .opencode/ — and reports, per item, the 5.0 kind and the
.agents/ path it belongs at. Keystone's own projections are skipped.

  (default)   report only; nothing is written
  --apply     write the converted files into .agents/. An existing .agents/
              file is never overwritten, and no original is changed.
  --overwrite --apply, then hand the originals to keystone: moved sections
              leave their root file, moved hooks leave settings.json, and
              each converted host file becomes keystone's, so the next
              projection replaces it with a link or shim (or removes it
              when no host reads that path any more). An original whose
              .agents/ target already existed is left alone.
`)
}
