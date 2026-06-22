package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// runIndex handles `keystone index [--dir <path>] [--harness-root <name>]`.
//
// Walks every canonical primitive location under .keystone/harness/
// (guides, actions, corpus, sensors, skills, agents, commands), parses
// each file's frontmatter, and writes .keystone/INDEX.json describing
// every primitive. The agent reads this artifact once at session start
// and opens bodies on demand.
func runIndex(args []string) error {
	dir := "."
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--help" || a == "-h":
			printIndexUsage(os.Stdout)
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
		fmt.Fprintf(os.Stderr, "keystone index: %s: %s\n", w.Path, w.Message)
	}

	idx := primitive.Build(primitives, time.Now())
	outPath := filepath.Join(absDir, config.KeystoneDir(harnessRoot), config.IndexName)
	if err := primitive.Write(outPath, idx); err != nil {
		return err
	}

	rel, _ := filepath.Rel(absDir, outPath)
	if rel == "" {
		rel = outPath
	}
	fmt.Fprintf(os.Stdout, "✓ keystone index — wrote %s (%d primitive(s) across %d kind(s))\n",
		rel, len(idx.Primitives), len(idx.ByKind))

	// INDEX.lite.json — the cheap discovery surface. CLAUDE.md points
	// agents here for first-pass browsing; the full INDEX is opened
	// only when a path/glob/trigger is needed to activate a primitive.
	lite := primitive.BuildLite(idx)
	liteOutPath := filepath.Join(absDir, config.KeystoneDir(harnessRoot), config.IndexLiteName)
	if err := primitive.WriteLite(liteOutPath, lite); err != nil {
		return err
	}
	liteRel, _ := filepath.Rel(absDir, liteOutPath)
	if liteRel == "" {
		liteRel = liteOutPath
	}
	fmt.Fprintf(os.Stdout, "✓ keystone index — wrote %s (lite descriptor)\n", liteRel)
	return nil
}

func printIndexUsage(w *os.File) {
	fmt.Fprint(w, `keystone index — emit the primitive descriptor index

Usage:
  keystone index [--dir <path>] [--harness-root <path>]

Walks the harness primitive locations and writes a single descriptor
artifact at <keystone-dir>/INDEX.json (one level above the harness
root — `+"`.keystone/INDEX.json`"+` for the default layout):

  Framework wrappers:
    <harness-root>/guides/**/*.md          → kind: guide      (wraps rule)
    <harness-root>/sensors/*.md            → kind: sensor     (wraps rule)
    <harness-root>/actions/*.md            → kind: action     (wraps command)
    <harness-root>/playbooks/*.md          → kind: playbook   (wraps skill)
    <harness-root>/personas/*.md           → kind: persona    (wraps subagent)

  Framework standalone:
    <harness-root>/corpus/**/*.md          → kind: corpus
    <harness-root>/evals/<id>/EVAL.md      → kind: eval
    <harness-root>/sources/*.md            → kind: source

  Agent escape hatches:
    <harness-root>/rules/*.md              → kind: rule
    <harness-root>/skills/<id>/SKILL.md    → kind: skill
    <harness-root>/agents/*.md             → kind: subagent
    <harness-root>/commands/*.md           → kind: command

Files without canonical frontmatter are skipped (pre-migration state);
files whose frontmatter fails to parse are reported on stderr and the
remaining primitives still index. README.md is never indexed.

The agent reads INDEX.json once at session start, then opens each
primitive's body only when the descriptor's globs / triggers / phase
match the work in hand.

Flags:
  --dir <path>           Project root (defaults to cwd).
  --harness-root <path>  Harness directory path (default: .keystone/harness).
`)
}
