package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/projection"
)

// runTarget dispatches `keystone target <subcommand> ...`.
func runTarget(args []string, _ fs.FS) error {
	if len(args) == 0 {
		printTargetUsage(os.Stderr)
		return fmt.Errorf("target requires a subcommand")
	}
	switch args[0] {
	case "add":
		return runTargetAdd(args[1:])
	case "help", "--help", "-h":
		printTargetUsage(os.Stdout)
		return nil
	default:
		return fmt.Errorf("unknown target subcommand %q (try: add)", args[0])
	}
}

func printTargetUsage(w *os.File) {
	fmt.Fprint(w, `keystone target — manage installed agent targets

Usage:
  keystone target add <agent>[,<agent>...] [--dir <path>]
  keystone target help

Commands:
  add    Install another agent target bundle into an existing charter.
`)
}

// runTargetAdd installs one or more additional agent target bundles into an
// existing charter directory. Errors out if any requested agent is already
// recorded in the lockfile — the user must explicitly remove it first
// rather than risk silent overwrites.
func runTargetAdd(args []string) error {
	dir := "."
	var positional []string

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--help" || a == "-h":
			printTargetAddUsage(os.Stdout)
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
			positional = append(positional, a)
		}
	}

	if len(positional) != 1 {
		return fmt.Errorf("target add requires exactly one agent argument (e.g. `keystone target add claude-code`); use --dir for the install directory")
	}
	rawAgents := positional[0]

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	charterRoot := config.DefaultAgentsRoot
	if _, err := os.Stat(filepath.Join(absDir, charterRoot)); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no %s/ in %s — run `keystone init` first", charterRoot, absDir)
		}
		return err
	}

	requested, err := parseTargetAgents(absDir, rawAgents)
	if err != nil {
		return err
	}

	existing, err := readInstalledAgents(absDir, charterRoot)
	if err != nil {
		return fmt.Errorf("read installed agents: %w", err)
	}
	already := map[string]bool{}
	for _, a := range existing {
		already[a] = true
	}
	for _, a := range requested {
		if already[a] {
			return fmt.Errorf("%s is already installed (recorded in the lockfile); remove it first to re-add", a)
		}
	}

	if err := appendInstalledAgents(absDir, charterRoot, requested); err != nil {
		return fmt.Errorf("update lockfile: %w", err)
	}
	if err := addAdapters(absDir, requested); err != nil {
		return fmt.Errorf("update keystone.json adapters: %w", err)
	}
	if err := runProject([]string{"--dir", absDir}); err != nil {
		return err
	}

	for _, agent := range requested {
		printAgentWarnings(agent)
	}
	printTargetAddNextSteps(requested, charterRoot)
	return nil
}

func printTargetAddUsage(w *os.File) {
	fmt.Fprint(w, `keystone target add — add another agent host to an existing charter

Usage:
  keystone target add <agent>[,<agent>...] [--dir <path>]

Requires .agents/ to exist in <dir> (default: .). Adds the agents to
keystone.json adapters and re-projects. Errors out if any of the
requested agents is already recorded in the lockfile.

Flags:
  --dir <path>   Project root containing .agents/ (defaults to cwd).
`)
}

// parseTargetAgents reads a comma-separated agent list: each a built-in
// host or one defined in .agents/hosts/, deduped, at least one.
func parseTargetAgents(absDir, raw string) ([]string, error) {
	var requested []string
	seen := map[string]bool{}
	for _, p := range strings.Split(raw, ",") {
		v := strings.TrimSpace(p)
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		if !isSupportedAgent(v) && !projection.IsCustomHost(absDir, v) {
			return nil, fmt.Errorf("unknown agent %q (supported: %v, or a host defined with `keystone new adapter`)", v, supportedAgents())
		}
		requested = append(requested, v)
	}
	if len(requested) == 0 {
		return nil, fmt.Errorf("no agents supplied")
	}
	return requested, nil
}

func printTargetAddNextSteps(agents []string, charterRoot string) {
	fmt.Fprintf(os.Stdout, "\n✓ added %s to the charter.\n\n", strings.Join(agents, ", "))
	fmt.Fprintln(os.Stdout, "  AGENTS.md says, per host, what it loads by itself and what to run by hand.")
	fmt.Fprintln(os.Stdout)
}
