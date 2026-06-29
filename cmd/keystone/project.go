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
	"github.com/tacoda/keystone/internal/framework/adapters/continueide"
	"github.com/tacoda/keystone/internal/framework/adapters/cursor"
	"github.com/tacoda/keystone/internal/framework/adapters/opencode"
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

	composed, composeErrs := primitive.Compose(primitives)
	for _, e := range composeErrs {
		fmt.Fprintf(os.Stderr, "keystone project: %s\n", e.Error())
	}
	primitives = composed

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

	// Run host adapters. Each adapter projects the host-specific
	// surface (hooks → .claude/settings.json, .cursor/rules/, etc.).
	// Source of truth for hooks lives in sensor frontmatter under
	// `.keystone/harness/sensors/` — the adapter reads the already-walked
	// primitive slice rather than walking again.
	cfg, cfgErr := config.ReadProjectConfig(absDir)
	if cfgErr != nil && !errors.Is(cfgErr, os.ErrNotExist) {
		fmt.Fprintf(os.Stderr, "keystone project: read keystone.json: %v\n", cfgErr)
	}

	// Agnostic surface: AGENTS.md is always projected so every coding
	// agent (Aider, Cline, Cursor, OpenCode, Roo, generic readers) gets
	// the same orientation Claude Code does.
	ares, err := agnostic.ProjectAgentsMD(absDir, agnostic.DefaultBody())
	if err != nil {
		return fmt.Errorf("agnostic AGENTS.md: %w", err)
	}
	if ares.Wrote {
		fmt.Fprintf(os.Stdout, "  wrote: %s\n", ares.Path)
	}

	// Claude Code adapter: hooks + posture regions in .claude/settings.json.
	if err := projectClaudeCode(absDir, primitives); err != nil {
		return err
	}

	// Opt-in cross-host adapters. Selection via keystone.json
	// `adapters:` list. Each adapter writes its host's surface and
	// reports the count of files emitted.
	if cfg != nil {
		if cfg.HasAdapter(config.AdapterCursor) {
			cres, err := cursor.ProjectRules(absDir, primitives)
			if err != nil {
				return fmt.Errorf("cursor rules: %w", err)
			}
			fmt.Fprintf(os.Stdout, "  cursor: %d rule(s) written, %d unchanged\n",
				cres.Wrote, cres.Unchanged)
		}
		if cfg.HasAdapter(config.AdapterAider) {
			ares, err := aider.ProjectAider(absDir, agnostic.DefaultBody())
			if err != nil {
				return fmt.Errorf("aider: %w", err)
			}
			if len(ares.Wrote) > 0 {
				fmt.Fprintf(os.Stdout, "  aider: wrote %s\n", strings.Join(ares.Wrote, ", "))
			} else {
				fmt.Fprintf(os.Stdout, "  aider: unchanged\n")
			}
		}
		if cfg.HasAdapter(config.AdapterContinue) {
			cnres, err := continueide.ProjectRules(absDir, primitives)
			if err != nil {
				return fmt.Errorf("continue rules: %w", err)
			}
			fmt.Fprintf(os.Stdout, "  continue: %d rule(s) written, %d unchanged\n",
				cnres.Wrote, cnres.Unchanged)
		}
		if cfg.HasAdapter(config.AdapterOpenCode) {
			ores, err := opencode.ProjectAgents(absDir, primitives)
			if err != nil {
				return fmt.Errorf("opencode: %w", err)
			}
			fmt.Fprintf(os.Stdout, "  opencode: %d file(s) written, %d unchanged (%d rule(s))\n",
				ores.Wrote, ores.Unchanged, ores.Rules)
		}
	}
	return nil
}

// projectClaudeCode emits the two Claude Code settings.json regions: the
// managed hooks block and the posture-derived permissions block.
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

Walks every primitive under .keystone/harness/ and writes host-native
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
target by design — collisions on the same id are caught by ` + "`keystone lint`" + `.

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
