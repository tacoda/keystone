package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/lockfile"
	"github.com/tacoda/keystone/internal/framework/plugins"
)

func runInit(args []string, assets fs.FS) error {
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

	harnessDest := filepath.Join(absDir, flags.harnessRoot)
	harnessExists := false
	if _, err := os.Stat(harnessDest); err == nil {
		harnessExists = true
	} else if !os.IsNotExist(err) {
		return err
	}

	mode := skipIfExists
	if flags.reset {
		if harnessExists {
			fmt.Fprintf(os.Stdout, "▸ --reset: removing existing %s/\n", harnessDest)
			if err := os.RemoveAll(harnessDest); err != nil {
				return fmt.Errorf("reset harness: %w", err)
			}
		}
		mode = overwrite
	} else if harnessExists {
		fmt.Fprintf(os.Stdout, "▸ %s/ already exists — writing only new files (existing files are kept; pass --reset to rewrite)\n", harnessDest)
	}

	fmt.Fprintf(os.Stdout, "▸ installing corpus to %s\n", harnessDest)
	if err := copyTree(assets, "harness", harnessDest, mode); err != nil {
		return fmt.Errorf("copy harness: %w", err)
	}

	for _, agent := range agents {
		if err := installAgentTarget(assets, agent, absDir); err != nil {
			return err
		}
	}

	if err := installConditional(assets, absDir, flags.harnessRoot, flags.selections); err != nil {
		return fmt.Errorf("install optional content: %w", err)
	}

	if err := initializeLockfile(absDir, flags.harnessRoot, agents); err != nil {
		return fmt.Errorf("initialize lockfile: %w", err)
	}

	if err := writeProjectConfigIfMissing(absDir, flags.harnessRoot); err != nil {
		return fmt.Errorf("initialize project config: %w", err)
	}

	if err := ensureGitignore(absDir, flags.harnessRoot); err != nil {
		return fmt.Errorf("update .gitignore: %w", err)
	}

	if err := writeInstallProfile(absDir, flags.harnessRoot, flags.selections, nil); err != nil {
		return fmt.Errorf("write install profile: %w", err)
	}

	for _, agent := range agents {
		printAgentWarnings(agent, flags.harnessRoot)
	}
	printNextSteps(agents, flags.harnessRoot)
	return nil
}

// initializeLockfile creates <harnessRoot>/keystone.lock.json if it doesn't
// exist yet, or fills in the keystone section if a prior call wrote only the
// policies section. Records the binary version, install date, the configured
// harness root, and the agent IDs whose menu files were just installed.
func initializeLockfile(destDir, harnessRoot string, agents []string) error {
	lf, err := ensureLockfile(destDir, harnessRoot)
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
	if err := lockfile.Write(destDir, harnessRoot, lf); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "  wrote: %s\n", filepath.Join(destDir, lockfile.RelPath(harnessRoot)))
	return nil
}

// installAgentTarget copies one agent's target bundle into destDir, merging its
// menu file. Existing target files are left alone (skipIfExists). If the agent
// has no shipped target bundle, prints a notice and returns nil — callers treat
// this as a soft success since the corpus alone is still useful.
func installAgentTarget(assets fs.FS, agent, destDir string) error {
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

// writeProjectConfigIfMissing creates keystone.json at the project root if
// it doesn't exist yet. Re-running init never overwrites a user-edited
// keystone.json — the user's plugin tree is preserved.
func writeProjectConfigIfMissing(projectDir, harnessRoot string) error {
	if _, err := config.ReadProjectConfig(projectDir); err == nil {
		fmt.Fprintf(os.Stdout, "  exists: %s (kept)\n", filepath.Join(projectDir, config.ProjectConfigFile))
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	cfg := config.DefaultProjectConfig(harnessRoot)
	if err := config.WriteProjectConfig(projectDir, cfg); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "  wrote: %s\n", filepath.Join(projectDir, config.ProjectConfigFile))
	return nil
}

// ensureGitignore makes sure .gitignore contains the vendored-plugins
// directory entry. Creates the file if missing; appends only the missing
// lines if present. The vendored tree is regenerated by `keystone install`
// from cache + sources, so it should not be tracked.
func ensureGitignore(projectDir, harnessRoot string) error {
	pluginsLine := filepath.ToSlash(filepath.Join(harnessRoot, plugins.PluginRoot)) + "/"
	path := filepath.Join(projectDir, ".gitignore")

	existing, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	body := string(existing)
	if alreadyHas(body, pluginsLine) {
		return nil
	}

	var b strings.Builder
	if len(body) > 0 {
		b.WriteString(body)
		if !strings.HasSuffix(body, "\n") {
			b.WriteByte('\n')
		}
		b.WriteByte('\n')
	}
	b.WriteString("# Keystone vendored plugins — regenerated by `keystone install`.\n")
	b.WriteString(pluginsLine)
	b.WriteByte('\n')

	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return err
	}
	verb := "updated"
	if len(existing) == 0 {
		verb = "wrote"
	}
	fmt.Fprintf(os.Stdout, "  %s: %s (added %s)\n", verb, path, pluginsLine)
	return nil
}

// alreadyHas reports whether line appears as its own line in body
// (ignoring leading/trailing whitespace).
func alreadyHas(body, line string) bool {
	for _, l := range strings.Split(body, "\n") {
		if strings.TrimSpace(l) == line {
			return true
		}
	}
	return false
}

func printNextSteps(agents []string, harnessRoot string) {
	list := strings.Join(agents, ", ")

	var bootstrapIn, lifecycle string
	if len(agents) == 1 {
		bootstrapIn = fmt.Sprintf("in %s", agents[0])
		lifecycle = fmt.Sprintf("     See %s/adapters/%s/lifecycle.md for how to invoke it.\n",
			harnessRoot, agentTargetDir(agents[0]))
	} else {
		bootstrapIn = "in any one of the agents above (the harness edits are agent-agnostic)"
		var b strings.Builder
		b.WriteString("     See:\n")
		for _, a := range agents {
			fmt.Fprintf(&b, "       %s/adapters/%s/lifecycle.md\n", harnessRoot, agentTargetDir(a))
		}
		b.WriteString("     for how to invoke it in each agent.\n")
		lifecycle = b.String()
	}

	fmt.Fprintf(os.Stdout, `
✓ harness installed for %s.

  ▶ Next: run the bootstrap action %s.

     Bootstrap reads your codebase and seeds %s/corpus/state/CODEBASE_STATE.md,
     %s/corpus/idioms/<your-stack>/, and the paired %s/guides/idioms/<your-stack>/
     so the harness reflects your project. It also confirms the sensor commands.
%s
Also:

  • Read %s/README.md for an overview of the five components
    (corpus, guides, sensors, policies, flywheels).
  • Review %s/corpus/state/INSTALL_PROFILE.md and adjust if needed.
  • Commit %s/ and any agent-specific files this installer created.
`, list, bootstrapIn, harnessRoot, harnessRoot, harnessRoot, lifecycle, harnessRoot, harnessRoot, harnessRoot)
}
