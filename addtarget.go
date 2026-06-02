package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// runAddTarget installs one or more additional agent target bundles into an
// existing harness directory. Errors out if any requested agent is already
// recorded in INSTALL_PROFILE.md — the user must explicitly remove it first
// rather than risk silent overwrites.
func runAddTarget(args []string, assets embed.FS) error {
	var dir string = "."
	var rawAgents string
	positional := []string{}

	for _, a := range args {
		switch {
		case a == "--help" || a == "-h":
			printAddTargetUsage(os.Stdout)
			return nil
		case strings.HasPrefix(a, "-"):
			return fmt.Errorf("unknown flag %s", a)
		default:
			positional = append(positional, a)
		}
	}

	switch len(positional) {
	case 0:
		return fmt.Errorf("add-target requires an agent name (e.g. `keystone add-target claude-code`)")
	case 1:
		rawAgents = positional[0]
	case 2:
		rawAgents = positional[0]
		dir = positional[1]
	default:
		return fmt.Errorf("add-target takes at most two positional arguments (<agent>[,<agent>...] [<dir>])")
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	if _, err := os.Stat(filepath.Join(absDir, "harness")); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no harness/ in %s — run `keystone init` first", absDir)
		}
		return err
	}

	// Parse and validate the requested agents.
	requested := []string{}
	seen := map[string]bool{}
	for _, p := range strings.Split(rawAgents, ",") {
		v := strings.TrimSpace(p)
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		if !isSupportedAgent(v) {
			return fmt.Errorf("unknown agent %q (supported: %v)", v, supportedAgents())
		}
		requested = append(requested, v)
	}
	if len(requested) == 0 {
		return fmt.Errorf("no agents supplied")
	}

	existing, err := readInstalledAgents(absDir)
	if err != nil {
		return fmt.Errorf("read install profile: %w", err)
	}
	already := map[string]bool{}
	for _, a := range existing {
		already[a] = true
	}
	for _, a := range requested {
		if already[a] {
			return fmt.Errorf("%s is already installed (recorded in INSTALL_PROFILE.md); remove it first to re-add", a)
		}
	}

	for _, agent := range requested {
		if err := installAgentTarget(assets, agent, absDir); err != nil {
			return err
		}
	}

	if err := appendAgentsToProfile(absDir, requested); err != nil {
		return fmt.Errorf("update install profile: %w", err)
	}

	for _, agent := range requested {
		printAgentWarnings(agent)
	}
	printAddTargetNextSteps(requested)
	return nil
}

func printAddTargetUsage(w *os.File) {
	fmt.Fprint(w, `keystone add-target — install another agent's target bundle into an existing harness

Usage:
  keystone add-target <agent>[,<agent>...] [<dir>]

Requires harness/ to exist in <dir> (default: .). Errors out if any of the
requested agents is already recorded in INSTALL_PROFILE.md.
`)
}

func printAddTargetNextSteps(agents []string) {
	fmt.Fprintf(os.Stdout, "\n✓ added %s to the harness.\n\n", strings.Join(agents, ", "))
	if len(agents) == 1 {
		fmt.Fprintf(os.Stdout, "  See harness/adapters/%s/lifecycle.md for how to invoke actions in it.\n",
			agentTargetDir(agents[0]))
	} else {
		fmt.Fprint(os.Stdout, "  See:\n")
		for _, a := range agents {
			fmt.Fprintf(os.Stdout, "    harness/adapters/%s/lifecycle.md\n", agentTargetDir(a))
		}
	}
	fmt.Fprintln(os.Stdout)
}
