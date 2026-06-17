package main

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/spf13/cobra"
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
	Use:           "keystone",
	Short:         "Install and operate the project harness",
	Long:          `keystone — install, audit, and serve the project harness from the CLI or MCP server.`,
	SilenceUsage:  true,
	SilenceErrors: true,
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
		Short:              "Scaffold the harness into a directory",
		Long:               "Scaffold the harness folder and the agent menu file(s) into <dir> (default: current directory).",
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
		RunE:               runAndForward(runVerify),
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
		Short:              "Manage agent targets installed under .keystone/harness/adapters/",
		DisableFlagParsing: true,
		RunE:               runAndForwardAssets(runTarget),
	}
	return c
}

func patchCmd() *cobra.Command {
	c := &cobra.Command{
		Use:                "patch [dir]",
		Short:              "Apply pending framework patches",
		DisableFlagParsing: true,
		RunE:               runAndForwardAssets(runPatch),
	}
	return c
}

func indexCmd() *cobra.Command {
	c := &cobra.Command{
		Use:                "index",
		Short:              "Emit .keystone/INDEX.json — primitive descriptors only",
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
