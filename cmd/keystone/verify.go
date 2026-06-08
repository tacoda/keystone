package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/loader"
	"github.com/tacoda/keystone/internal/framework/lockfile"
	"github.com/tacoda/keystone/internal/framework/plugins"
)

// runVerify handles `keystone verify [--dir <path>] [--harness-root <name>]`.
//
// Walks the plugin tree from keystone.json, checks each vendored plugin
// for drift against the lockfile, and reports any strict-cascade
// violations from project files shadowing locked items. Exits non-zero
// on any violation; drift alone exits zero but is reported and reset.
func runVerify(args []string) error {
	flagValue, args, err := extractHarnessRootFlag(args)
	if err != nil {
		return err
	}
	dir := "."
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--help" || a == "-h":
			printVerifyUsage(os.Stdout)
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
	harnessRoot, err := resolveHarnessRoot(absDir, flagValue)
	if err != nil {
		return err
	}

	cfg, err := config.ReadProjectConfig(absDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no %s at %s — run `keystone init` first", config.ProjectConfigFile, absDir)
		}
		return err
	}

	lf, err := lockfile.Read(absDir, harnessRoot)
	if err != nil {
		return err
	}
	expected := map[string]map[string]string{}
	for name, lock := range lf.Plugins {
		expected[name] = lock.Files
	}

	res, err := loader.Verify(absDir, cfg, expected)
	if err != nil {
		return err
	}

	if res.HasDrift() {
		fmt.Fprintf(os.Stdout, "▸ drift detected — resetting %d plugin(s)\n", len(res.Drift))
		for _, d := range res.Drift {
			fmt.Fprintf(os.Stdout, "  • %s: %d drifted file(s)\n", d.Plugin, len(d.Files))
			for _, f := range d.Files {
				fmt.Fprintf(os.Stdout, "      - %s (%s)\n", f.Path, f.Kind)
			}
			if err := plugins.Reset(d.Plugin, absDir, harnessRoot); err != nil {
				return fmt.Errorf("reset %s: %w", d.Plugin, err)
			}
		}
		fmt.Fprintln(os.Stdout, "  re-run `keystone install` to repopulate from cache")
	}

	if res.HasErrors() {
		fmt.Fprintf(os.Stdout, "✗ keystone verify found %d strict violation(s) in %s\n\n", len(res.Violations), absDir)
		for _, v := range res.Violations {
			fmt.Fprintln(os.Stdout, "  "+v.String())
			fmt.Fprintln(os.Stdout)
		}
		fmt.Fprintln(os.Stdout, "Strict plugin items cannot be overridden by the project layer. Remove the offending file(s) or take it up with the plugin author.")
		return fmt.Errorf("strict cascade is violated")
	}

	if !res.HasDrift() {
		fmt.Fprintf(os.Stdout, "✓ keystone verify clean — no drift, no strict violations (%s)\n", absDir)
	}
	return nil
}

func printVerifyUsage(w *os.File) {
	fmt.Fprint(w, `keystone verify — check vendored plugins and the strict cascade

Usage:
  keystone verify [--dir <path>] [--harness-root <name>]

Reads keystone.json + the lockfile, then:
  - Walks every vendored plugin and compares per-file hashes to the
    lockfile. Any drift triggers an immediate plugins.Reset (run
    `+"`keystone install`"+` afterward to repopulate).
  - Walks each plugin's strict items and reports project-layer files
    that shadow them (e.g. <harness-root>/<port>/<item>.md present when
    a plugin marks <port>/<item> strict).

Exit codes:
  0  clean (no drift after reset, no violations)
  0  drift only — drifted plugins were reset; user re-installs to recover
  1  any strict violation

Flags:
  --dir <path>           Project root (defaults to cwd).
  --harness-root <name>  Harness directory name.
`)
}
