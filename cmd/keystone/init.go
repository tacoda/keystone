package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tacoda/keystone/internal/framework/lockfile"
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

	installedPolicies, err := installPolicies(absDir, flags.policies)
	if err != nil {
		return fmt.Errorf("install policies: %w", err)
	}

	if err := initializeLockfile(absDir, agents); err != nil {
		return fmt.Errorf("initialize lockfile: %w", err)
	}

	if len(installedPolicies) > 0 {
		res, verr := verifyPolicies(absDir)
		if verr != nil {
			return fmt.Errorf("policy verify: %w", verr)
		}
		if printVerifyReport(absDir, res) {
			return fmt.Errorf("init succeeded but strict policy cascade is violated — resolve the shadowing file(s) above")
		}
	}

	if err := writeInstallProfile(absDir, flags.selections, installedPolicies); err != nil {
		return fmt.Errorf("write install profile: %w", err)
	}

	for _, agent := range agents {
		printAgentWarnings(agent)
	}
	printNextSteps(agents)
	return nil
}

// initializeLockfile creates harness/.keystone.lock if it doesn't exist yet,
// or fills in the keystone section if a prior call (e.g. installPacks) wrote
// only the packs section. Records the binary version, install date, and the
// agent IDs whose menu files were just installed.
func initializeLockfile(destDir string, agents []string) error {
	lf, err := ensureLockfile(destDir)
	if err != nil {
		return err
	}
	lf.Keystone.Version = version
	if lf.Keystone.Installed == "" {
		lf.Keystone.Installed = time.Now().UTC().Format("2006-01-02")
	}
	seen := map[string]bool{}
	for _, a := range lf.Keystone.Agents {
		seen[a] = true
	}
	for _, a := range agents {
		if !seen[a] {
			lf.Keystone.Agents = append(lf.Keystone.Agents, a)
			seen[a] = true
		}
	}
	if err := lockfile.Write(destDir, lf); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "  wrote: %s\n", filepath.Join(destDir, lockfile.File))
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

  • Read harness/README.md for an overview of the five components
    (corpus, guides, sensors, policies, flywheels).
  • Review harness/corpus/state/INSTALL_PROFILE.md and adjust if needed.
  • Commit harness/ and any agent-specific files this installer created.
`, list, bootstrapIn, lifecycle)
}
