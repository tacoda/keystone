package main

import (
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tacoda/keystone/internal/framework/migrations"
)

// rootCmd is the keystone CLI's top-level Cobra command. Subcommands
// are added in init() functions in adjacent files so each command's
// declaration sits next to its handler.
//
// Each command's RunE delegates to the existing run<Verb> function and
// passes the raw positional args through unchanged. Local flag parsing
// stays inside the run functions for now; a later pass will lift flags
// to Cobra-native definitions.
var rootCmd = &cobra.Command{
	Use:               "keystone",
	Short:             "Install and operate the project charter",
	Long:              `keystone — install, audit, and serve the project charter from the CLI or MCP server.`,
	SilenceUsage:      true,
	SilenceErrors:     true,
	PersistentPreRunE: warnIfPendingMigrations,
}

// warnIfPendingMigrations prints a one-line stderr warning when the
// install at the current working directory has migrations the registry
// knows about but that the lockfile hasn't recorded as applied. Never
// returns an error — the warning is advisory; commands proceed normally
// (per the migrations no-breaking-changes invariant).
//
// Skips migrate / options / version / help / completion to avoid
// confusing or duplicate output during the commands that already speak
// to this state.
func warnIfPendingMigrations(cmd *cobra.Command, _ []string) error {
	switch cmd.Name() {
	case "migrate", "options", "version", "help", "completion", "keystone":
		return nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}
	applied := readAppliedTolerant(cwd)
	pending := migrations.Pending(applied)
	if len(pending) == 0 {
		return nil
	}
	var versions []string
	for _, m := range pending {
		versions = append(versions, m.Version)
	}
	fmt.Fprintf(os.Stderr, "⚠ %d pending migration(s): %s — run `keystone migrate up` to apply\n",
		len(pending), strings.Join(versions, ", "))
	return nil
}

// assetsFS holds the embedded scaffold templates. Set from main(); kept
// at package scope so command RunE functions can reach it without
// threading it through every constructor.
var assetsFS fs.FS

// SetAssets is called from main to inject the embedded scaffold tree
// before rootCmd.Execute runs.
func SetAssets(assets fs.FS) {
	assetsFS = assets
}

func init() {
	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(installCmd())
	rootCmd.AddCommand(policyCmd())
	rootCmd.AddCommand(verifyCmd())
	rootCmd.AddCommand(doctorCmd())
	rootCmd.AddCommand(newCmd())
	rootCmd.AddCommand(targetCmd())
	rootCmd.AddCommand(patchCmd())
	rootCmd.AddCommand(indexCmd())
	rootCmd.AddCommand(lintCmd())
	rootCmd.AddCommand(listCmd())
	rootCmd.AddCommand(showCmd())
	rootCmd.AddCommand(documentCmd())
	rootCmd.AddCommand(hookCmd())
	rootCmd.AddCommand(signalCmd())
	rootCmd.AddCommand(charterCmd())
	rootCmd.AddCommand(explainCmd())
	rootCmd.AddCommand(projectCmd())
	rootCmd.AddCommand(migrateCmd())
	rootCmd.AddCommand(mcpCmd())
	rootCmd.AddCommand(webCmd())
	rootCmd.AddCommand(evalCmd())
	rootCmd.AddCommand(searchCmd())
	rootCmd.AddCommand(graphCmd())
	rootCmd.AddCommand(watchCmd())
	rootCmd.AddCommand(snapshotCmd())
	rootCmd.AddCommand(optionsCmd())
	rootCmd.AddCommand(versionCmd())
}

// runAndForward is the standard RunE shim — passes the args slice on
// to a hand-rolled run function, and propagates errors so Cobra prints
// them with the appropriate exit code.
func runAndForward(fn func(args []string) error) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, args []string) error {
		if err := fn(args); err != nil {
			return err
		}
		return nil
	}
}

func runAndForwardAssets(fn func(args []string, assets fs.FS) error) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, args []string) error {
		if err := fn(args, assetsFS); err != nil {
			return err
		}
		return nil
	}
}

// initCmd, installCmd, etc. live in their per-command files.

func initCmd() *cobra.Command {
	c := &cobra.Command{
		Use:                "init [dir]",
		Short:              "Scaffold the charter into a directory",
		Long:               "Scaffold the charter folder and the agent menu file(s) into <dir> (default: current directory).",
		DisableFlagParsing: true,
		RunE:               runAndForwardAssets(runInit),
	}
	return c
}

func installCmd() *cobra.Command {
	c := &cobra.Command{
		Use:                "install",
		Short:              "Materialize every policy declared in keystone.json",
		DisableFlagParsing: true,
		RunE:               runAndForward(runInstall),
	}
	return c
}

func policyCmd() *cobra.Command {
	c := &cobra.Command{
		Use:                "policy",
		Short:              "Manage installed policies (add, update, remove)",
		DisableFlagParsing: true,
		RunE:               runAndForward(runPolicy),
	}
	return c
}

func verifyCmd() *cobra.Command {
	c := &cobra.Command{
		Use:                "verify",
		Short:              "Check vendored policies for drift and the strict cascade for violations",
		DisableFlagParsing: true,
		RunE:               runAndForward(verifyWithHooks),
	}
	return c
}

func doctorCmd() *cobra.Command {
	c := &cobra.Command{
		Use:                "doctor",
		Short:              "Audit the install — path conventions, policy integrity, template drift",
		DisableFlagParsing: true,
		RunE:               runAndForward(runDoctor),
	}
	return c
}

func newCmd() *cobra.Command {
	c := &cobra.Command{
		Use:                "new <kind> <id>",
		Short:              "Scaffold a new primitive at the conventional path",
		DisableFlagParsing: true,
		RunE:               runAndForward(runNew),
	}
	return c
}

func targetCmd() *cobra.Command {
	c := &cobra.Command{
		Use:                "target",
		Short:              "Manage agent targets installed under .charter/adapters/",
		DisableFlagParsing: true,
		RunE:               runAndForwardAssets(runTarget),
	}
	return c
}

// patchCmd is a hidden stub. The patches subsystem was retired in 2.1
// and replaced by `keystone migrate`. Kept so users on older muscle
// memory get a friendly redirect instead of "unknown command".
func patchCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "patch [args...]",
		Short:              "Retired in 2.1 — use `keystone migrate` instead",
		Hidden:             true,
		DisableFlagParsing: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Fprintln(os.Stdout, "⚠ `keystone patch` was retired in 2.1.")
			fmt.Fprintln(os.Stdout, "  Use `keystone migrate` for forward and backward version transforms:")
			fmt.Fprintln(os.Stdout, "    keystone migrate up      apply pending migrations")
			fmt.Fprintln(os.Stdout, "    keystone migrate down    roll back the most recent migration")
			fmt.Fprintln(os.Stdout, "    keystone migrate status  show applied + pending")
			return nil
		},
	}
}

func indexCmd() *cobra.Command {
	c := &cobra.Command{
		Use:                "index",
		Short:              "Emit .charter/INDEX.json — primitive descriptors only",
		DisableFlagParsing: true,
		RunE:               runAndForward(runIndex),
	}
	return c
}

func lintCmd() *cobra.Command {
	c := &cobra.Command{
		Use:                "lint",
		Short:              "Validate primitive frontmatter (required fields, unique ids, deps)",
		DisableFlagParsing: true,
		RunE:               runAndForward(runLint),
	}
	return c
}

func listCmd() *cobra.Command {
	c := &cobra.Command{
		Use:                "list [<kind>] [--tag <tag>]...",
		Short:              "List primitives, filtered by kind and / or tag",
		DisableFlagParsing: true,
		RunE:               runAndForward(runList),
	}
	return c
}

func showCmd() *cobra.Command {
	c := &cobra.Command{
		Use:                "show <kind> <id>",
		Short:              "Show one primitive's descriptor + cross-references",
		DisableFlagParsing: true,
		RunE:               runAndForward(runShow),
	}
	return c
}

func documentCmd() *cobra.Command {
	c := &cobra.Command{
		Use:                "document",
		Short:              "List document instances and advance them through gates",
		DisableFlagParsing: true,
		RunE:               runAndForward(runDocument),
	}
	return c
}

func hookCmd() *cobra.Command {
	c := &cobra.Command{
		Use:                "hook",
		Short:              "Fire framework hooks bound to a workflow event",
		DisableFlagParsing: true,
		RunE:               runAndForward(runHook),
	}
	return c
}

func signalCmd() *cobra.Command {
	c := &cobra.Command{
		Use:                "signal",
		Short:              "Fire or list keystone signals (extensible framework events)",
		DisableFlagParsing: true,
		RunE:               runAndForward(runSignal),
	}
	return c
}

func charterCmd() *cobra.Command {
	c := &cobra.Command{
		Use:                "charter",
		Short:              "Inspect the charter (coverage: files no guide governs)",
		DisableFlagParsing: true,
		RunE:               runAndForward(runCharter),
	}
	return c
}

func explainCmd() *cobra.Command {
	c := &cobra.Command{
		Use:                "explain",
		Short:              "Explain a primitive — how it activates, what it links to, changes",
		DisableFlagParsing: true,
		RunE:               runAndForward(runExplain),
	}
	return c
}

func projectCmd() *cobra.Command {
	c := &cobra.Command{
		Use:                "project",
		Short:              "Regenerate host-native projections (.claude/...) from canonical sources",
		DisableFlagParsing: true,
		RunE:               runAndForward(runProject),
	}
	return c
}

func migrateCmd() *cobra.Command {
	c := &cobra.Command{
		Use:                "migrate",
		Short:              "One-shot 1.x → 2.0 layout + schema upgrade",
		DisableFlagParsing: true,
		RunE:               runAndForward(runMigrate),
	}
	return c
}

func optionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "options",
		Short: "Print the allowed labels for every option flag",
		Run: func(*cobra.Command, []string) {
			printOptionLabels(os.Stdout)
		},
	}
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   "Print the binary version",
		Aliases: []string{"-v"},
		Run: func(*cobra.Command, []string) {
			fmt.Println(version)
		},
	}
}
