package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tacoda/keystone/internal/framework/charter"
	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// runExplain handles `keystone explain <id> [--kind K] [--dir D]`. It
// explains a primitive — what it is, how it activates, what it links to,
// where it projects — and flags whether it has uncommitted changes.
func runExplain(args []string) error {
	id, kind, dir := parseExplainArgs(args)
	if id == "" {
		printExplainUsage(os.Stderr)
		return fmt.Errorf("`keystone explain` requires a primitive <id>")
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	prims, _, err := primitive.Walk(absDir, config.DefaultCharterRoot)
	if err != nil {
		return fmt.Errorf("walk charter: %w", err)
	}
	matches := matchByID(prims, id, kind)
	switch len(matches) {
	case 0:
		return fmt.Errorf("no primitive with id %q%s", id, kindSuffix(kind))
	case 1:
		printExplanation(charter.Explain(matches[0]), gitDirty(absDir, matches[0].Path))
		return nil
	default:
		return ambiguousErr(id, matches)
	}
}

func parseExplainArgs(args []string) (id, kind, dir string) {
	dir = "."
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--kind":
			if i+1 < len(args) {
				kind = args[i+1]
				i++
			}
		case "--dir":
			if i+1 < len(args) {
				dir = args[i+1]
				i++
			}
		default:
			if id == "" && !strings.HasPrefix(args[i], "-") {
				id = args[i]
			}
		}
	}
	return id, kind, dir
}

func matchByID(prims []primitive.Primitive, id, kind string) []primitive.Primitive {
	var out []primitive.Primitive
	for _, p := range prims {
		if charter.Matches(p, id, kind) {
			out = append(out, p)
		}
	}
	return out
}

func ambiguousErr(id string, matches []primitive.Primitive) error {
	kinds := make([]string, 0, len(matches))
	for _, m := range matches {
		kinds = append(kinds, m.Kind)
	}
	return fmt.Errorf("id %q matches multiple kinds (%s) — narrow with --kind", id, strings.Join(kinds, ", "))
}

func kindSuffix(kind string) string {
	if kind == "" {
		return ""
	}
	return " of kind " + kind
}

// gitDirty reports whether the file at relPath has uncommitted changes
// (modified or untracked). Best-effort: false if git is unavailable.
func gitDirty(projectDir, relPath string) bool {
	cmd := exec.Command("git", "-C", projectDir, "status", "--porcelain", "--", relPath)
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}

func printExplanation(e charter.Explanation, dirty bool) {
	src := ""
	if e.Provenance != "" && e.Provenance != "project" {
		src = "  [" + e.Provenance + "]"
	}
	fmt.Fprintf(os.Stdout, "%s  %s%s\n", e.Kind, e.ID, src)
	fmt.Fprintf(os.Stdout, "  %s\n\n", e.Description)
	fmt.Fprintf(os.Stdout, "  Activation:  %s\n", e.Activation)
	if len(e.Links) > 0 {
		fmt.Fprintf(os.Stdout, "  Links:       %s\n", strings.Join(e.Links, "; "))
	}
	if e.ProjectsTo != "" {
		fmt.Fprintf(os.Stdout, "  Projects to: %s (via keystone project)\n", e.ProjectsTo)
	}
	fmt.Fprintf(os.Stdout, "  Body:        %s\n", e.BodyPath)
	if dirty {
		fmt.Fprintf(os.Stdout, "\n  ⚠ uncommitted changes — this primitive differs from the last commit.\n")
	}
}

func printExplainUsage(w *os.File) {
	fmt.Fprint(w, `keystone explain — explain a primitive

Usage:

    keystone explain <id> [--kind K] [--dir D]

Explains a primitive: what it is, how/when it activates (a guide's globs,
a sensor's on:, a skill's triggers, …), what it links to, and where it
projects. Flags whether the primitive has uncommitted changes.
`)
}
