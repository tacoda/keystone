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

// runIndex handles `keystone index [--dir <path>] [--charter-root <name>]`.
//
// Walks every canonical primitive location under .agents/
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
	charterRoot := config.DefaultAgentsRoot

	primitives, warnings, err := primitive.Walk(absDir, charterRoot)
	if err != nil {
		return err
	}
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "keystone index: %s: %s\n", w.Path, w.Message)
	}

	// Resolve concern `includes:` before indexing so the descriptor
	// records the merged shape the agent actually sees.
	composed, composeErrs := primitive.Compose(primitives)
	for _, e := range composeErrs {
		fmt.Fprintf(os.Stderr, "keystone index: %s\n", e.Error())
	}
	primitives = composed

	idx := primitive.Build(primitives, time.Now())
	outPath := filepath.Join(absDir, config.StateRoot, config.IndexName)
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
	liteOutPath := filepath.Join(absDir, config.StateRoot, config.IndexLiteName)
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
  keystone index [--dir <path>]

Walks .agents/ and writes .keystone/INDEX.json (every descriptor) and
.keystone/INDEX.lite.json (category, kind, id, description). A
primitive's kind comes from its folder:

  Constraints:
    .agents/rules/**/*.md              → rule
    .agents/references/**/*.md         → reference
    .agents/hooks/*.md                 → hook
    .agents/permissions/*.md           → permission
    .agents/evals/<id>/EVAL.md         → eval

  Capabilities:
    .agents/skills/<id>/SKILL.md       → skill
    .agents/commands/*.md              → command
    .agents/agents/*.md                → subagent
    .agents/tools/*.md                 → tool
    .agents/documents/*.md             → document

Vendored policies under .keystone/policies/ are indexed the same way.
The 4.x folders (guides/, corpus/, sensors/, playbooks/, …) still index
under their 5.0 kinds. Files without frontmatter are skipped; files whose
frontmatter fails to parse are reported on stderr. README.md is never
indexed.

The agent reads the index first, then opens a body only when the
primitive applies (a rule's paths:, a skill's description, a hook's
event).

Flags:
  --dir <path>    Project root (defaults to cwd).
`)
}
