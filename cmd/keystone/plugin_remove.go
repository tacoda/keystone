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

// runPluginRemove handles `keystone plugin remove <name> [--dir <path>] [--harness-root <name>]`.
//
// Removes the named plugin from keystone.json, deletes its vendor
// directory, and drops its entry from the lockfile. Re-adding it later
// fetches fresh from the source.
func runPluginRemove(args []string) error {
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
			printPluginRemoveUsage(os.Stdout)
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
		return fmt.Errorf("plugin remove requires exactly one plugin name")
	}
	name := positional[0]

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
			return fmt.Errorf("no %s at %s — nothing to remove", config.ProjectConfigFile, absDir)
		}
		return err
	}

	pruned, removed := removePlugin(cfg.Plugins, name)
	if !removed {
		return fmt.Errorf("plugin %q is not declared in %s", name, config.ProjectConfigFile)
	}
	cfg.Plugins = pruned

	if err := config.WriteProjectConfig(absDir, cfg); err != nil {
		return err
	}

	if err := plugins.Reset(name, absDir, harnessRoot); err != nil {
		return err
	}

	lf, err := lockfile.Read(absDir, harnessRoot)
	if err != nil {
		return err
	}
	delete(lf.Plugins, name)
	if err := lockfile.Write(absDir, harnessRoot, lf); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "✓ removed %s\n", name)
	return nil
}

func printPluginRemoveUsage(w *os.File) {
	fmt.Fprint(w, `keystone plugin remove — uninstall a plugin

Usage:
  keystone plugin remove <name> [--dir <path>] [--harness-root <name>]

Removes the named plugin from keystone.json, deletes its vendor directory,
and drops its lockfile entry.

Flags:
  --dir <path>           Project root (defaults to cwd).
  --harness-root <name>  Harness directory name.
`)
}
