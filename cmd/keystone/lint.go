package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// runLint handles `keystone lint [--dir <path>]`.
//
// Walks every primitive under .agents/, parses frontmatter,
// runs Lint(), prints findings, and exits non-zero on any error-level
// finding. Warnings are reported but never block the exit.
func runLint(args []string) error {
	opts, help, err := parseLintArgs(args)
	if err != nil || help {
		return err
	}
	dir, verbose, fix := opts.dir, opts.verbose, opts.fix

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
		fmt.Fprintf(os.Stderr, "keystone lint: %s: %s\n", w.Path, w.Message)
	}
	if fix {
		if primitives, err = applyLintFixes(absDir, charterRoot, primitives); err != nil {
			return err
		}
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

type lintOpts struct {
	dir          string
	verbose, fix bool
}

// parseLintArgs reads `keystone lint` flags; help=true means usage was printed.
func parseLintArgs(args []string) (opts lintOpts, help bool, err error) {
	opts.dir = "."
	help, err = parseFlags(args, map[string]*string{"--dir": &opts.dir},
		map[string]*bool{"--verbose": &opts.verbose, "-v": &opts.verbose, "--fix": &opts.fix},
		func() { printLintUsage(os.Stdout) })
	return opts, help, err
}

// applyLintFixes runs the --fix rewrites, then re-walks so the findings
// below describe the fixed tree.
func applyLintFixes(absDir, charterRoot string, primitives []primitive.Primitive) ([]primitive.Primitive, error) {
	n, err := fixFrontmatter(absDir, primitives)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(os.Stdout, "▸ keystone lint --fix rewrote %d file(s)\n", n)
	fixed, _, err := primitive.Walk(absDir, charterRoot)
	return fixed, err
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
  keystone lint [--dir <path>] [--verbose] [--fix]

Walks every canonical primitive under .agents/, checks each
file's frontmatter against the canonical schema, and reports any
violations. Exits non-zero on any error-level finding.

Hard errors (block exit):
  - missing kind / id / description
  - unknown kind value
  - duplicate (kind, id) across the charter
  - empty paths entry
  - subagent missing tools
  - SKILL.md name invalid or not matching its directory; description
    over 1024 or compatibility over 500 characters (Agent Skills spec)

Warnings (reported with --verbose; never block):
  - description still says "TODO"
  - id contains characters outside [A-Za-z0-9-:_/.]
  - deps[] / traces[] entries that do not resolve
  - SKILL.md without name, or with keys outside the Agent Skills spec
  - concern (retired in 5.0)

Flags:
  --dir <path>    Project root (defaults to cwd).
  --verbose, -v   Show warnings in addition to errors.
  --fix           Rewrite project SKILL.md frontmatter into Agent Skills
                  spec shape (name, triggers folded into description,
                  tools → allowed-tools, model/tags → metadata), and
                  inline retired concerns (includes:) into their
                  consumers. Bodies are untouched; every rewrite is printed.
`)
}
