package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/lockfile"
	"github.com/tacoda/keystone/internal/framework/plugins"
)

// runPluginUpdate handles `keystone plugin update <name> [@<new-version>]
// [--dir <path>] [--harness-root <name>]`.
//
// Looks up the named plugin in keystone.json, optionally bumps its version,
// then resets and re-installs it. Re-fetching is the right move even when
// the recorded ref is unchanged because that ref may be moving (a branch
// like `main` picks up new commits).
func runPluginUpdate(args []string) error {
	flagValue, args, err := extractHarnessRootFlag(args)
	if err != nil {
		return err
	}
	dir := "."
	var positional []string

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--help" || a == "-h":
			printPluginUpdateUsage(os.Stdout)
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

	if len(positional) == 0 {
		return fmt.Errorf("plugin update requires a plugin name (e.g. `keystone plugin update tacoda-org`)")
	}
	if len(positional) > 2 {
		return fmt.Errorf("plugin update takes at most two positional arguments (<name> [@<new-version>])")
	}

	name := positional[0]
	var newVersion string
	if len(positional) == 2 {
		newVersion = strings.TrimPrefix(positional[1], "@")
		if newVersion == "" {
			return fmt.Errorf("explicit version cannot be empty; use just `<name>` to re-fetch the recorded ref")
		}
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	harnessRoot, err := resolveHarnessRoot(absDir, flagValue)
	if err != nil {
		return err
	}

	cfg, err := config.ReadProjectConfig(absDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no %s at %s — run `keystone init` and `keystone plugin add` first", config.ProjectConfigFile, absDir)
		}
		return err
	}

	node := findPlugin(cfg.Plugins, name)
	if node == nil {
		return fmt.Errorf("plugin %q is not declared in %s", name, config.ProjectConfigFile)
	}
	if newVersion != "" {
		node.Version = newVersion
	}

	if err := config.WriteProjectConfig(absDir, cfg); err != nil {
		return err
	}

	// Always reset before reinstall — picks up any moving-ref drift even
	// when the version field hasn't changed.
	if err := plugins.Reset(node.Name, absDir, harnessRoot); err != nil {
		return err
	}

	lf, err := lockfile.Read(absDir, harnessRoot)
	if err != nil {
		return err
	}
	if err := installOnePlugin(absDir, harnessRoot, *node, lf); err != nil {
		return err
	}
	if err := lockfile.Write(absDir, harnessRoot, lf); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "✓ updated %s → %s @ %s\n", node.Name, node.Source, node.Version)
	return nil
}

func printPluginUpdateUsage(w *os.File) {
	fmt.Fprint(w, `keystone plugin update — re-resolve and reinstall an installed plugin

Usage:
  keystone plugin update <name> [@<new-version>] [--dir <path>] [--harness-root <name>]

If <new-version> is supplied, the version field in keystone.json is bumped
and the new ref is fetched. Without a version, the recorded ref is re-
fetched — useful for moving refs like a branch.

Either way the vendor directory is reset and reinstalled from cache, so
any local edits are discarded.

Flags:
  --dir <path>           Project root (defaults to cwd).
  --harness-root <name>  Harness directory name.
`)
}
