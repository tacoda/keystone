package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// runProject handles `keystone project [--dir <path>]`.
//
// Walks every primitive under .keystone/harness/ and regenerates the
// host-native projections under .claude/ from the canonical sources.
// Hand-edits to projections are overwritten; the drift sensor flags
// them as findings.
func runProject(args []string) error {
	dir := "."
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--help" || a == "-h":
			printProjectUsage(os.Stdout)
			return nil
		case a == "--dir":
			if i+1 >= len(args) {
				return fmt.Errorf("flag %s requires a value", a)
			}
			dir = args[i+1]
			i++
		case strings.HasPrefix(a, "--dir="):
			dir = strings.TrimPrefix(a, "--dir=")
		case strings.HasPrefix(a, "-"):
			return fmt.Errorf("unknown flag %s", a)
		default:
			return fmt.Errorf("unexpected positional argument %q", a)
		}
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	harnessRoot := config.DefaultHarnessRoot

	primitives, warnings, err := primitive.Walk(absDir, harnessRoot)
	if err != nil {
		return err
	}
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "keystone project: %s: %s\n", w.Path, w.Message)
	}

	results, err := primitive.Project(absDir, primitives)
	if err != nil {
		return err
	}

	wrote := 0
	for _, r := range results {
		if r.Action == "wrote" {
			fmt.Fprintf(os.Stdout, "  wrote: %s\n", r.Dest)
			wrote++
		}
	}
	fmt.Fprintf(os.Stdout, "✓ keystone project — projected %d primitive(s)\n", wrote)
	return nil
}

func printProjectUsage(w *os.File) {
	fmt.Fprint(w, `keystone project — regenerate host-native projections

Usage:
  keystone project [--dir <path>]

Walks every primitive under .keystone/harness/ and writes host-native
projections from the canonical sources:

  kind: skill    → .claude/skills/<id>/SKILL.md
  kind: subagent → .claude/agents/<id>.md
  kind: command  → .claude/commands/<id>.md

Disk-name normalization: ids containing ":" (canonical namespace
separator) are rewritten to "-" in the projection filename; the
frontmatter id stays unchanged.

Hand-edits to projection files are erased on the next run — the
canonical source under .keystone/harness/ is the only file you edit.
The drift sensor reports projections that diverge from their source.

Flags:
  --dir <path>    Project root (defaults to cwd).
`)
}
