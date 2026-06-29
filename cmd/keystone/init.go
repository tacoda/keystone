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
	"github.com/tacoda/keystone/internal/framework/policies"
	"github.com/tacoda/keystone/internal/framework/primitive"
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
	if err := ensureProjectDir(absDir); err != nil {
		return err
	}

	if err := resolveAgent(flags, absDir); err != nil {
		return err
	}

	// 2.0 init is minimum-friction. The agent target is the ONLY thing
	// init asks about — everything else (app type, architecture,
	// testing, compliance, starter packs) is detected by the bootstrap
	// action against the actual codebase, or asked there as
	// context-only questions ("what's your aspirational pattern?")
	// the codebase can't answer for you.
	//
	// Resolution order for the agent:
	//   1. --agent <name> flag (highest priority)
	//   2. marker-file detection (resolveAgent above)
	//   3. one stdin prompt if TTY
	//   4. fall back to generic (non-TTY without flag + no markers)
	if _, ok := flags.selections["agent"]; !ok {
		if isTerminal(os.Stdin) {
			pick, err := promptAgent(os.Stdin, os.Stdout)
			if err != nil {
				return err
			}
			if pick != "" {
				flags.selections["agent"] = []string{pick}
			}
		}
		if _, ok := flags.selections["agent"]; !ok {
			fmt.Fprintln(os.Stderr, "! no --agent provided; falling back to `generic` target")
			flags.selections["agent"] = []string{"generic"}
		}
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

	fmt.Fprintf(os.Stdout, "▸ installing harness to %s\n", harnessDest)
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

	if err := writeInstallProfile(absDir, flags.harnessRoot, flags.selections); err != nil {
		return fmt.Errorf("write install profile: %w", err)
	}

	// Index + project: write .keystone/INDEX.json so the agent can read
	// the primitive descriptor at session start, and project skills /
	// subagents / commands into the host-native tree (`.claude/skills/`,
	// `.claude/agents/`, `.claude/commands/`) so they're discoverable as
	// slash commands. Without this step a fresh install ships skills the
	// host agent can't see. Errors are non-fatal — the structural install
	// is already complete and these artifacts can be regenerated with
	// `keystone index` / `keystone project`.
	if err := indexAndProject(absDir, flags.harnessRoot); err != nil {
		fmt.Fprintf(os.Stderr, "! post-install index/project skipped: %v\n", err)
	}

	if err := reportAmbientLoad(absDir, flags.harnessRoot); err != nil {
		// Non-fatal: budget reporting is a nicety, not a precondition.
		fmt.Fprintf(os.Stderr, "! ambient-load report skipped: %v\n", err)
	}

	for _, agent := range agents {
		printAgentWarnings(agent, flags.harnessRoot)
	}
	printNextSteps(agents, flags.harnessRoot)
	return nil
}

// reportAmbientLoad prints a one-line per-port token count after a fresh
// install. Gives the user an at-a-glance sense of how heavy the install
// is before they start adding policies or custom content. Honors
// keystone.json's budgets block (just-written by init), so users with a
// pre-existing budgets section see the over-budget warning surfaced on
// the first run too.
func reportAmbientLoad(projectDir, harnessRoot string) error {
	alloc, err := walkHarnessBudget(projectDir, harnessRoot)
	if err != nil {
		return err
	}
	cfg, _ := config.ReadProjectConfig(projectDir)
	reps := alloc.Report(cfg, 0)
	if len(reps) == 0 {
		return nil
	}
	total := 0
	for _, r := range reps {
		total += r.Tokens
	}
	fmt.Fprintf(os.Stdout, "\n▸ ambient load: %d tokens across %d port(s)\n", total, len(reps))
	for _, r := range reps {
		suffix := ""
		if r.MaxTokens > 0 {
			if r.IsOverBudget() {
				suffix = fmt.Sprintf(" (over budget by %d)", r.OverBy)
			} else {
				suffix = fmt.Sprintf(" (cap %d, %d%% used)", r.MaxTokens, 100*r.Tokens/r.MaxTokens)
			}
		}
		fmt.Fprintf(os.Stdout, "    %-10s %d%s\n", r.Port, r.Tokens, suffix)
	}
	fmt.Fprintln(os.Stdout, "  (see the localhost dashboard via `keystone web serve` for per-port detail)")
	return nil
}

// indexAndProject regenerates .keystone/INDEX.json and projects every
// agent-kind primitive (skill / subagent / command) into the host-native
// tree under .claude/. Called once at the end of `keystone init` so a
// fresh install ships discoverable slash commands without an extra
// manual step.
func indexAndProject(projectDir, harnessRoot string) error {
	primitives, warnings, err := primitive.Walk(projectDir, harnessRoot)
	if err != nil {
		return fmt.Errorf("walk primitives: %w", err)
	}
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "  ! %s: %s\n", w.Path, w.Message)
	}

	idx := primitive.Build(primitives, time.Now())
	indexPath := filepath.Join(projectDir, config.KeystoneDir(harnessRoot), config.IndexName)
	if err := primitive.Write(indexPath, idx); err != nil {
		return fmt.Errorf("write index: %w", err)
	}
	if rel, e := filepath.Rel(projectDir, indexPath); e == nil {
		fmt.Fprintf(os.Stdout, "  wrote: %s (%d primitive(s))\n", rel, len(primitives))
	}

	results, err := primitive.Project(projectDir, primitives)
	if err != nil {
		return fmt.Errorf("project primitives: %w", err)
	}
	wrote := 0
	for _, r := range results {
		if r.Action == "wrote" {
			if rel, e := filepath.Rel(projectDir, r.Dest); e == nil {
				fmt.Fprintf(os.Stdout, "  wrote: %s\n", rel)
			}
			wrote++
		}
	}
	if wrote > 0 {
		fmt.Fprintf(os.Stdout, "  projected %d primitive(s) to host-native paths\n", wrote)
	}
	return nil
}

// initializeLockfile creates <harnessRoot>/keystone.lock.json if it doesn't
// exist yet, or fills in the keystone section if a prior call only wrote
// policy entries. Records the binary version, install date, the configured
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
// ensureProjectDir creates the target directory if absent (init scaffolds
// into a fresh dir the same as an existing one) and rejects a non-dir path.
func ensureProjectDir(absDir string) error {
	info, err := os.Stat(absDir)
	if os.IsNotExist(err) {
		return os.MkdirAll(absDir, 0o755)
	}
	if err != nil {
		return fmt.Errorf("dir %s: %w", absDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("dir %s is not a directory", absDir)
	}
	return nil
}

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
// keystone.json — the user's policy tree is preserved.
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

// ensureGitignore makes sure .gitignore contains the vendored-policies
// directory entry. Creates the file if missing; appends only the missing
// lines if present. The vendored tree is regenerated by `keystone install`
// from cache + sources, so it should not be tracked.
func ensureGitignore(projectDir, harnessRoot string) error {
	policiesLine := filepath.ToSlash(filepath.Join(harnessRoot, policies.PolicyRoot)) + "/"
	path := filepath.Join(projectDir, ".gitignore")

	existing, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	body := string(existing)
	if alreadyHas(body, policiesLine) {
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
	b.WriteString("# Keystone vendored policies — regenerated by `keystone install`.\n")
	b.WriteString(policiesLine)
	b.WriteByte('\n')

	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return err
	}
	verb := "updated"
	if len(existing) == 0 {
		verb = "wrote"
	}
	fmt.Fprintf(os.Stdout, "  %s: %s (added %s)\n", verb, path, policiesLine)
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

  ▶ Next: run the /keystone:bootstrap skill %s.

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
