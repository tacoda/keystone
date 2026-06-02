package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func runInit(args []string, assets embed.FS) error {
	flags, err := parseInitArgs(args)
	if err != nil {
		return err
	}

	absDir, err := filepath.Abs(flags.dir)
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	if info, err := os.Stat(absDir); err != nil {
		return fmt.Errorf("dir %s: %w", absDir, err)
	} else if !info.IsDir() {
		return fmt.Errorf("dir %s is not a directory", absDir)
	}

	if err := resolveAgent(flags, absDir); err != nil {
		return err
	}

	// Categories already satisfied by CLI flags or by agent detection are
	// skipped during prompting. Everything else is asked for if we're in a TTY.
	skip := map[string]bool{}
	for id, v := range flags.selections {
		if len(v) > 0 {
			skip[id] = true
		}
	}

	if tty := isTerminal(os.Stdin); tty {
		if err := promptMissing(flags.selections, skip); err != nil {
			return err
		}
	} else {
		// Non-TTY: agent is mandatory; everything else can stay unset.
		if _, ok := flags.selections["agent"]; !ok {
			return fmt.Errorf("no --agent provided, detection found nothing, and stdin is not a TTY for interactive prompts")
		}
		fmt.Fprintf(os.Stderr, "! non-interactive mode: optional categories left unset\n")
	}

	agents := flags.selections["agent"]
	for _, a := range agents {
		if !isSupportedAgent(a) {
			return fmt.Errorf("unknown agent %q (supported: %v)", a, supportedAgents())
		}
	}

	preflight(absDir)

	harnessDest := filepath.Join(absDir, "harness")
	if _, err := os.Stat(harnessDest); err == nil {
		if !flags.force {
			return fmt.Errorf("%s already exists; pass --force to overwrite", harnessDest)
		}
		fmt.Fprintf(os.Stdout, "▸ overwriting existing harness/\n")
	} else if !os.IsNotExist(err) {
		return err
	}

	fmt.Fprintf(os.Stdout, "▸ installing corpus to %s\n", harnessDest)
	if err := copyTree(assets, "harness", harnessDest, overwrite); err != nil {
		return fmt.Errorf("copy harness: %w", err)
	}

	for _, agent := range agents {
		if err := installAgentTarget(assets, agent, absDir); err != nil {
			return err
		}
	}

	if err := installConditional(assets, absDir, flags.selections); err != nil {
		return fmt.Errorf("install optional content: %w", err)
	}

	if err := writeInstallProfile(absDir, flags.selections); err != nil {
		return fmt.Errorf("write install profile: %w", err)
	}

	for _, agent := range agents {
		printAgentWarnings(agent)
	}
	printNextSteps(agents)
	return nil
}

// installAgentTarget copies one agent's target bundle into destDir, merging its
// menu file. Existing target files are left alone (skipIfExists). If the agent
// has no shipped target bundle, prints a notice and returns nil — callers treat
// this as a soft success since the corpus alone is still useful.
func installAgentTarget(assets embed.FS, agent, destDir string) error {
	targetRoot := filepath.Join("targets", agentTargetDir(agent))
	if _, err := assets.Open(targetRoot); err != nil {
		fmt.Fprintf(os.Stderr, "! no target bundle for %s; configure activation manually\n", agent)
		return nil
	}
	fmt.Fprintf(os.Stdout, "▸ installing %s target\n", agent)
	menuPath, err := installMenuFile(assets, agent, destDir)
	if err != nil {
		return fmt.Errorf("install menu file for %s: %w", agent, err)
	}
	if err := copyTree(assets, targetRoot, destDir, skipIfExists, menuPath); err != nil {
		return fmt.Errorf("copy target for %s: %w", agent, err)
	}
	return nil
}

// resolveAgent fills flags.selections["agent"] when it is not already set,
// using existing marker-file detection. It does NOT prompt — that is left
// to promptMissing, which runs after this step.
func resolveAgent(flags *initFlags, dir string) error {
	if _, set := flags.selections["agent"]; set {
		return nil
	}
	detected := detectAgent(dir)
	if detected == "" {
		return nil
	}
	fmt.Fprintf(os.Stdout, "▸ detected agent: %s\n", detected)
	flags.selections["agent"] = []string{detected}
	return nil
}

func preflight(dir string) {
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		fmt.Fprintf(os.Stderr, "! %s is not a git repository; keystone assumes git-tracked projects\n", dir)
	}

	ciCandidates := []string{
		".github/workflows",
		".gitlab-ci.yml",
		".circleci/config.yml",
		".travis.yml",
		"azure-pipelines.yml",
		"bitbucket-pipelines.yml",
		"Jenkinsfile",
	}
	for _, c := range ciCandidates {
		if _, err := os.Stat(filepath.Join(dir, c)); err == nil {
			return
		}
	}
	fmt.Fprintf(os.Stderr, "! no CI config detected; the release phase assumes a CI pipeline\n")
}

func printNextSteps(agents []string) {
	list := strings.Join(agents, ", ")

	var bootstrapIn, lifecycle string
	if len(agents) == 1 {
		bootstrapIn = fmt.Sprintf("in %s", agents[0])
		lifecycle = fmt.Sprintf("     See harness/adapters/%s/lifecycle.md for how to invoke it.\n",
			agentTargetDir(agents[0]))
	} else {
		bootstrapIn = "in any one of the agents above (the harness edits are agent-agnostic)"
		var b strings.Builder
		b.WriteString("     See:\n")
		for _, a := range agents {
			fmt.Fprintf(&b, "       harness/adapters/%s/lifecycle.md\n", agentTargetDir(a))
		}
		b.WriteString("     for how to invoke it in each agent.\n")
		lifecycle = b.String()
	}

	fmt.Fprintf(os.Stdout, `
✓ harness installed for %s.

  ▶ Next: run the bootstrap action %s.

     Bootstrap reads your codebase and seeds harness/corpus/state/CODEBASE_STATE.md,
     harness/corpus/idioms/<your-stack>/, and the paired harness/guides/idioms/<your-stack>/
     so the harness reflects your project. It also confirms the sensor commands.
%s
Also:

  • Read harness/README.md for an overview of the four components
    (corpus, guides, sensors, flywheels).
  • Review harness/corpus/state/INSTALL_PROFILE.md and adjust if needed.
  • Commit harness/ and any agent-specific files this installer created.
`, list, bootstrapIn, lifecycle)
}
