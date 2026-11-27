package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/drift"
)

func driftCmd() *cobra.Command {
	c := &cobra.Command{
		Use:                "drift [files…|--diff] [--json]",
		Short:              "Context-drift report: dead links, vanished paths:, stale verified: dates (warnings only)",
		DisableFlagParsing: true,
		RunE:               runAndForward(runDrift),
	}
	return c
}

// runDrift handles `keystone drift [files…|--diff] [--json] [--dir D]`.
// The report, human + --json. `files…` scopes the scan to primitives
// relevant to those files (paths: match or the primitive's own file);
// `--diff` takes the scope from `git diff --name-only HEAD`. Warnings
// only — keystone names stale context, it never judges the diff.
func runDrift(args []string) error {
	var files []string
	asJSON, diffMode := false, false
	for i := 0; i < len(args); i++ {
		switch a := args[i]; {
		case a == "--json":
			asJSON = true
		case a == "--diff":
			diffMode = true
		case a == "--dir":
			i++
			if i >= len(args) {
				return fmt.Errorf("--dir requires a value")
			}
		case strings.HasPrefix(a, "--dir="):
		case strings.HasPrefix(a, "-"):
			return fmt.Errorf("unknown flag %q", a)
		default:
			files = append(files, a)
		}
	}
	absDir, err := filepath.Abs(dirArg(args))
	if err != nil {
		return err
	}
	if diffMode {
		files, err = gitDiffFiles(absDir)
		if err != nil {
			return err
		}
	}
	days := config.DefaultDriftDays
	if cfg, err := config.ReadProjectConfig(absDir); err == nil {
		days = cfg.DriftWindow()
	}
	rep, err := drift.Scan(absDir, drift.Options{Files: files, Days: days})
	if err != nil {
		return err
	}
	if asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(rep)
	}
	if len(rep.Warnings) == 0 {
		fmt.Fprintf(os.Stdout, "✓ no context drift (checked %d primitives)\n", rep.Checked)
		return nil
	}
	fmt.Fprintf(os.Stdout, "⚠ %d drift warning(s) (checked %d primitives):\n", len(rep.Warnings), rep.Checked)
	for _, w := range rep.Warnings {
		fmt.Fprintf(os.Stdout, "  • %s — %s (%s/%s %s)\n", w.Path, w.Reason, w.Kind, w.PrimID, w.Detail)
	}
	return nil
}

// gitDiffFiles lists the repo-relative files changed vs HEAD (staged and
// unstaged), for `keystone drift --diff`.
func gitDiffFiles(absDir string) ([]string, error) {
	cmd := exec.Command("git", "-C", absDir, "diff", "--name-only", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff --name-only HEAD: %w", err)
	}
	var files []string
	for _, line := range strings.Split(string(out), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			files = append(files, filepath.ToSlash(line))
		}
	}
	return files, nil
}
