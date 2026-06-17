package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// runLint handles `keystone lint [--dir <path>]`.
//
// Walks every primitive under .keystone/harness/, parses frontmatter,
// runs Lint(), prints findings, and exits non-zero on any error-level
// finding. Warnings are reported but never block the exit.
func runLint(args []string) error {
	dir := "."
	verbose := false
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--help" || a == "-h":
			printLintUsage(os.Stdout)
			return nil
		case a == "--dir":
			if i+1 >= len(args) {
				return fmt.Errorf("flag %s requires a value", a)
			}
			dir = args[i+1]
			i++
		case strings.HasPrefix(a, "--dir="):
			dir = strings.TrimPrefix(a, "--dir=")
		case a == "--verbose" || a == "-v":
			verbose = true
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
		fmt.Fprintf(os.Stderr, "keystone lint: %s: %s\n", w.Path, w.Message)
	}

	findings := primitive.Lint(primitives)

	errCount, warnCount := 0, 0
	for _, f := range findings {
		switch f.Severity {
		case primitive.FindingError:
			errCount++
		case primitive.FindingWarning:
			warnCount++
		}
		if verbose || f.Severity == primitive.FindingError {
			fmt.Fprintln(os.Stdout, "  "+f.String())
		}
	}

	if errCount == 0 && warnCount == 0 {
		fmt.Fprintf(os.Stdout, "✓ keystone lint clean — %d primitive(s) across %d kind(s)\n",
			len(primitives), kindCount(primitives))
		return nil
	}
	if errCount == 0 {
		fmt.Fprintf(os.Stdout, "✓ keystone lint clean (errors=0, warnings=%d). Re-run with --verbose to see warnings.\n", warnCount)
		return nil
	}
	fmt.Fprintf(os.Stderr, "✗ keystone lint failed — %d error(s), %d warning(s)\n", errCount, warnCount)
	return fmt.Errorf("lint failed")
}

func kindCount(ps []primitive.Primitive) int {
	seen := map[string]bool{}
	for _, p := range ps {
		seen[p.Kind] = true
	}
	return len(seen)
}

func printLintUsage(w *os.File) {
	fmt.Fprint(w, `keystone lint — validate primitive frontmatter

Usage:
  keystone lint [--dir <path>] [--verbose]

Walks every canonical primitive under .keystone/harness/, checks each
file's frontmatter against the canonical schema, and reports any
violations. Exits non-zero on any error-level finding.

Hard errors (block exit):
  - missing kind / id / description
  - unknown kind value
  - duplicate (kind, id) across the harness
  - empty globs entry
  - skill missing triggers, subagent missing tools

Warnings (reported with --verbose; never block):
  - description still says "TODO"
  - id contains characters outside [A-Za-z0-9-:_/.]
  - deps[] / traces[] entries that do not resolve

Flags:
  --dir <path>    Project root (defaults to cwd).
  --verbose, -v   Show warnings in addition to errors.
`)
}
