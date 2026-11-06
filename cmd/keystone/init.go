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
	"github.com/tacoda/keystone/internal/framework/migrations"
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

	charterDest := filepath.Join(absDir, flags.charterRoot)
	charterExists := false
	if _, err := os.Stat(charterDest); err == nil {
		charterExists = true
	} else if !os.IsNotExist(err) {
		return err
	}

	// Never delete: --overwrite replaces keystone's own scaffold files in
	// place; files the user added stay.
	mode := skipIfExists
	if flags.overwrite {
		mode = overwrite
	} else if charterExists {
		fmt.Fprintf(os.Stdout, "▸ %s/ already exists — writing only new files (existing files are kept; pass --overwrite to replace keystone's)\n", charterDest)
	}

	if err := installScaffold(assets, absDir, charterDest, mode); err != nil {
		return err
	}

	if err := installConditional(assets, absDir, flags.charterRoot, flags.selections); err != nil {
		return fmt.Errorf("install optional content: %w", err)
	}

	if err := initializeLockfile(absDir, flags.charterRoot, agents); err != nil {
		return fmt.Errorf("initialize lockfile: %w", err)
	}

	if err := writeProjectConfigIfMissing(absDir, agents); err != nil {
		return fmt.Errorf("initialize project config: %w", err)
	}

	if err := ensureGitignore(absDir); err != nil {
		return fmt.Errorf("update .gitignore: %w", err)
	}

	if err := writeInstallProfile(absDir, flags.charterRoot, flags.selections); err != nil {
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
	if err := indexAndProject(absDir, flags.charterRoot, flags.overwrite); err != nil {
		fmt.Fprintf(os.Stderr, "! post-install index/project skipped: %v\n", err)
	}

	if err := reportAmbientLoad(absDir, flags.charterRoot); err != nil {
		// Non-fatal: budget reporting is a nicety, not a precondition.
		fmt.Fprintf(os.Stderr, "! ambient-load report skipped: %v\n", err)
	}

	for _, agent := range agents {
		printAgentWarnings(agent)
	}
	printNextSteps(agents, flags.charterRoot)
	return nil
}

// reportAmbientLoad prints, after a fresh install, what each host loads
// every session, so the user sees how heavy the install is (and any cap a
// pre-existing budgets block sets) before adding content.
func reportAmbientLoad(projectDir, _ string) error {
	r, err := measureBudget(projectDir)
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout)
	printBudgetSummary(os.Stdout, r)
	return nil
}

// indexAndProject regenerates .keystone/INDEX.json and projects every
// agent-kind primitive (skill / subagent / command) into the host-native
// tree under .claude/. Called once at the end of `keystone init` so a
// fresh install ships discoverable slash commands without an extra
// manual step.
func indexAndProject(projectDir, charterRoot string, overwrite bool) error {
	primitives, warnings, err := primitive.Walk(projectDir, charterRoot)
	if err != nil {
		return fmt.Errorf("walk primitives: %w", err)
	}
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "  ! %s: %s\n", w.Path, w.Message)
	}

	idx := primitive.Build(primitives, time.Now())
	indexPath := filepath.Join(projectDir, config.StateRoot, config.IndexName)
	if err := primitive.Write(indexPath, idx); err != nil {
		return fmt.Errorf("write index: %w", err)
	}
	if rel, e := filepath.Rel(projectDir, indexPath); e == nil {
		fmt.Fprintf(os.Stdout, "  wrote: %s (%d primitive(s))\n", rel, len(primitives))
	}

	if _, err := projectHosts(projectDir, primitives, os.Stdout, overwrite); err != nil {
		return fmt.Errorf("project primitives: %w", err)
	}

	// Root entrypoints (AGENTS.md, CLAUDE.md, .aider.conf.yml) are part of
	// the projection above.
	return nil
}

// initializeLockfile creates <charterRoot>/keystone.lock.json if it doesn't
// exist yet, or fills in the keystone section if a prior call only wrote
// policy entries. Records the binary version, install date, the configured
// charter root, and the agent IDs whose menu files were just installed.
// markFreshInstallMigrated records every migration as applied on a brand
// new install: it starts on the current layout, so nothing is pending. An
// older layout (.charter/, .harness/) in the folder means a real upgrade
// is still to run, so nothing is recorded.
func markFreshInstallMigrated(destDir string, lf *lockfile.Lockfile) {
	for _, legacy := range []string{".charter", ".harness"} {
		if _, err := os.Stat(filepath.Join(destDir, legacy)); err == nil {
			return
		}
	}
	lf.MigrationsApplied = nil
	for _, m := range migrations.All() {
		lf.MigrationsApplied = append(lf.MigrationsApplied, m.Version)
	}
}

func initializeLockfile(destDir, charterRoot string, agents []string) error {
	lf, err := ensureLockfile(destDir, charterRoot)
	if err != nil {
		return err
	}
	lf.Keystone.Version = version
	if lf.Keystone.Installed == "" {
		lf.Keystone.Installed = time.Now().UTC().Format("2006-01-02")
		markFreshInstallMigrated(destDir, lf)
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
	fmt.Fprintf(os.Stdout, "  wrote: %s\n", filepath.Join(destDir, lockfile.RelPath()))
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
// The chosen agents become `adapters`, which selects the hosts `keystone
// project` writes for; on an existing keystone.json they are only added.
func writeProjectConfigIfMissing(projectDir string, agents []string) error {
	if _, err := config.ReadProjectConfig(projectDir); err == nil {
		fmt.Fprintf(os.Stdout, "  exists: %s (kept)\n", filepath.Join(projectDir, config.ProjectConfigFile))
		return addAdapters(projectDir, agents)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	cfg := config.DefaultProjectConfig("")
	cfg.Adapters = hostAdapters(agents)
	if err := config.WriteProjectConfig(projectDir, cfg); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "  wrote: %s\n", filepath.Join(projectDir, config.ProjectConfigFile))
	return nil
}

// hostAdapters maps `--agent` choices to keystone.json adapter names. The
// generic choice needs no adapter: every project gets AGENTS.md.
func hostAdapters(agents []string) []string {
	var out []string
	for _, a := range agents {
		if a != "generic" && a != "_generic" {
			out = append(out, a)
		}
	}
	return out
}

// addAdapters appends agents missing from keystone.json `adapters`; the
// user's existing entries and order are kept.
func addAdapters(projectDir string, agents []string) error {
	cfg, err := config.ReadProjectConfig(projectDir)
	if err != nil {
		return err
	}
	changed := false
	for _, a := range hostAdapters(agents) {
		if !cfg.HasAdapter(a) {
			cfg.Adapters = append(cfg.Adapters, a)
			changed = true
		}
	}
	if !changed {
		return nil
	}
	return config.WriteProjectConfig(projectDir, cfg)
}

// installScaffold copies the two scaffold trees: templates/agents → the
// charter's primitives root, templates/keystone → .keystone/ state.
func installScaffold(assets fs.FS, absDir, charterDest string, mode copyMode) error {
	fmt.Fprintf(os.Stdout, "▸ installing charter to %s\n", charterDest)
	if err := copyTree(assets, "agents", charterDest, mode); err != nil {
		return fmt.Errorf("copy charter: %w", err)
	}
	stateDest := filepath.Join(absDir, config.StateRoot)
	fmt.Fprintf(os.Stdout, "▸ installing keystone state to %s\n", stateDest)
	if err := copyTree(assets, "keystone", stateDest, mode); err != nil {
		return fmt.Errorf("copy keystone state: %w", err)
	}
	return nil
}

// ensureGitignore makes sure .gitignore contains the vendored-policies
// directory entry. Creates the file if missing; appends only the missing
// lines if present. The vendored tree is regenerated by `keystone install`
// from cache + sources, so it should not be tracked.
func ensureGitignore(projectDir string) error {
	policiesLine := config.PoliciesDir + "/"
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

func printNextSteps(agents []string, charterRoot string) {
	list := strings.Join(agents, ", ")

	var bootstrapIn, lifecycle string
	if len(agents) == 1 {
		bootstrapIn = fmt.Sprintf("in %s", agents[0])
		lifecycle = "     AGENTS.md says what that host loads by itself and what to run by hand.\n"
	} else {
		bootstrapIn = "in any one of the agents above (the charter edits are agent-agnostic)"
		lifecycle = "     AGENTS.md says, per host, what it loads by itself and what to run by hand.\n"
	}

	fmt.Fprintf(os.Stdout, `
✓ charter installed for %s.

  ▶ Next: run the /keystone:bootstrap skill %s.

     Bootstrap reads your codebase and seeds %s/references/state/CODEBASE_STATE.md,
     %s/rules/idioms/<your-stack>/, and the paired %s/references/idioms/<your-stack>/
     so the charter reflects your project. It also confirms the hook commands.
%s
Also:

  • Read %s/README.md for an overview: constraints (rules, references,
    hooks, permissions, evals) and capabilities (skills, commands,
    subagents, tools, documents).
  • Review %s/references/state/INSTALL_PROFILE.md and adjust if needed.
  • Commit %s/ and any agent-specific files this installer created.
`, list, bootstrapIn, charterRoot, charterRoot, charterRoot, lifecycle, charterRoot, charterRoot, charterRoot)
}
