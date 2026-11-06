package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tacoda/keystone/internal/framework/adapters/agnostic"
	"github.com/tacoda/keystone/internal/framework/adapters/aider"
	"github.com/tacoda/keystone/internal/framework/adapters/claudecode"
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

	// Run host adapters. Each adapter projects the host-specific
	// surface (hooks → .claude/settings.json, .cursor/rules/, etc.).
	// Source of truth for hooks lives in sensor frontmatter under
	// `.agents/hooks/` — the adapter reads the already-walked
	// primitive slice rather than walking again.
	cfg, cfgErr := config.ReadProjectConfig(absDir)
	if cfgErr != nil && !errors.Is(cfgErr, os.ErrNotExist) {
		fmt.Fprintf(os.Stderr, "keystone project: read keystone.json: %v\n", cfgErr)
	}

	// Canonical entrypoint (CHARTER.md) + the always-on AGENTS.md
	// pointer. Every other host file is a thin pointer to CHARTER.md.
	if err := writeCharterSurface(absDir, true); err != nil {
		return err
	}

	// Claude Code adapter: hooks + posture regions in .claude/settings.json.
	if err := projectClaudeCode(absDir, primitives); err != nil {
		return err
	}

	// Opt-in cross-host adapters. Selection via keystone.json
	// `adapters:` list. Each adapter writes its host's surface and
	// reports the count of files emitted.
	if cfg != nil {
		if cfg.HasAdapter(config.AdapterAider) {
			ares, err := aider.ProjectAider(absDir, agnostic.RenderPointer(agnostic.AiderProfile()))
			if err != nil {
				return fmt.Errorf("aider: %w", err)
			}
			if len(ares.Wrote) > 0 {
				fmt.Fprintf(os.Stdout, "  aider: wrote %s\n", strings.Join(ares.Wrote, ", "))
			} else {
				fmt.Fprintf(os.Stdout, "  aider: unchanged\n")
			}
		}
	}
	return nil
}

// projectClaudeCode emits the two Claude Code settings.json regions: the
// managed hooks block and the posture-derived permissions block.
// writeCharterSurface emits the always-on entrypoint surface: the
// canonical CHARTER.md plus the thin AGENTS.md pointer to it. Shared by
// `keystone project`, `keystone init`, and the watch loop so the three
// never drift. When verbose, prints a "wrote:" line per changed file.
func writeCharterSurface(projectDir string, verbose bool) error {
	cres, err := agnostic.ProjectCharterMD(projectDir)
	if err != nil {
		return fmt.Errorf("CHARTER.md: %w", err)
	}
	ares, err := agnostic.ProjectAgentsMD(projectDir, agnostic.RenderPointer(agnostic.GenericProfile()))
	if err != nil {
		return fmt.Errorf("agnostic AGENTS.md: %w", err)
	}
	if verbose {
		for _, r := range []agnostic.ProjectionResult{cres, ares} {
			if r.Wrote {
				fmt.Fprintf(os.Stdout, "  wrote: %s\n", r.Path)
			}
		}
	}
	return nil
}

func projectClaudeCode(absDir string, primitives []primitive.Primitive) error {
	hres, err := claudecode.ProjectHooks(absDir, primitives)
	if err != nil {
		return fmt.Errorf("claudecode hooks: %w", err)
	}
	if hres.Wrote {
		fmt.Fprintf(os.Stdout, "  wrote: %s (+%d managed hook(s), -%d stale)\n",
			hres.Path, hres.Added, hres.Removed)
	} else if hres.Added > 0 {
		fmt.Fprintf(os.Stdout, "  unchanged: %s (%d managed hook(s))\n", hres.Path, hres.Added)
	}

	pres, err := claudecode.ProjectPosture(absDir, primitives)
	if err != nil {
		return fmt.Errorf("claudecode posture: %w", err)
	}
	if pres.Wrote {
		fmt.Fprintf(os.Stdout, "  wrote: %s (+%d permission(s))\n", pres.Path, pres.Added)
	}
	return nil
}

func printProjectUsage(w *os.File) {
	fmt.Fprint(w, `keystone project — regenerate host-native projections

Usage:
  keystone project [--dir <path>]

Walks every primitive under .agents/ and writes host-native
projections from the canonical sources:

  Framework wrappers (encouraged authoring path):
    kind: persona  → .claude/agents/<id>.md
    kind: action   → .claude/commands/<id>.md
    kind: playbook → .claude/skills/<id>/SKILL.md

  Agent escape hatches (raw host-native, same targets):
    kind: subagent → .claude/agents/<id>.md
    kind: command  → .claude/commands/<id>.md
    kind: skill    → .claude/skills/<id>/SKILL.md

A framework wrapper and its agent counterpart share the same .claude/
target by design — collisions on the same id are caught by `+"`keystone lint`"+`.

Disk-name normalization: ids containing ":" (canonical namespace
separator) are rewritten to "-" in the projection filename; the
frontmatter id stays unchanged.

Hand-edits to projection files are erased on the next run — the
canonical source under .agents/ is the only file you edit.
The drift sensor reports projections that diverge from their source.

Flags:
  --dir <path>    Project root (defaults to cwd).
`)
}
