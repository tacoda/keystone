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
// Walks every primitive under .agents/ and regenerates the
// host-native projections under .claude/ from the canonical sources.
// Hand-edits to projections are overwritten; the drift sensor flags
// them as findings.
func runProject(args []string) error {
	dir := "."
	overwrite := false
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
		case a == "--overwrite":
			overwrite = true
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
	charterRoot := config.DefaultAgentsRoot

	primitives, warnings, err := primitive.Walk(absDir, charterRoot)
	if err != nil {
		return err
	}
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "keystone project: %s: %s\n", w.Path, w.Message)
	}

	composed, composeErrs := primitive.Compose(primitives)
	for _, e := range composeErrs {
		fmt.Fprintf(os.Stderr, "keystone project: %s\n", e.Error())
	}
	primitives = composed

	res, err := projectHosts(absDir, primitives, os.Stdout, overwrite)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "✓ keystone project — %d written, %d unchanged, %d pruned\n",
		len(res.Wrote), len(res.Unchanged), len(res.Pruned))

	return nil
}

func printProjectUsage(w *os.File) {
	fmt.Fprint(w, `keystone project — regenerate host-native projections

Usage:
  keystone project [--dir <path>] [--overwrite]

Delivers .agents/ to every host in keystone.json "adapters" (and any
custom host defined in .agents/hosts/), by the cheapest method each host
supports:

  native   the host already reads the path (Codex and opencode read
           .agents/skills/)
  link     a symlink where the host reads the same format elsewhere
           (Claude Code: .claude/skills/<name>); "link_mode": "copy"
           writes copies instead
  shim     a translated file where the host needs another format
           (.claude/agents/<id>.md, .cursor/rules/<slug>.mdc,
           .github/instructions/<slug>.instructions.md, …)
  index    an AGENTS.md entry where the host has no slot for the kind

It also writes the AGENTS.md and CLAUDE.md managed blocks, the hook
bridges, and the permission rules each host can enforce, and names each
gap a host can't close.

Keystone records every file it writes in .keystone/lockfile.json and
never replaces a file it didn't write; --overwrite replaces it. Edit the
source in .agents/, never a projected file: `+"`keystone verify`"+`
reports projections that drift from their source.

Flags:
  --dir <path>    Project root (defaults to cwd).
  --overwrite     Replace host files keystone didn't write.
`)
}
