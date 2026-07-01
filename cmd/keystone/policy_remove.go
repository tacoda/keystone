package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/lockfile"
	"github.com/tacoda/keystone/internal/framework/policies"
)

// runPolicyRemove handles `keystone policy remove <name> [--dir <path>] [--charter-root <name>]`.
//
// Removes the named policy from keystone.json, deletes its vendor
// directory, and drops its entry from the lockfile. Re-adding it later
// fetches fresh from the source.
func runPolicyRemove(args []string) error {
	dir := "."
	var positional []string

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--help" || a == "-h":
			printPolicyRemoveUsage(os.Stdout)
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
		return fmt.Errorf("policy remove requires exactly one policy name")
	}
	name := positional[0]

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	charterRoot := config.DefaultCharterRoot

	cfg, err := config.ReadProjectConfig(absDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no %s at %s — nothing to remove", config.ProjectConfigFile, absDir)
		}
		return err
	}

	pruned, removed := removePolicy(cfg.Policies, name)
	if !removed {
		return fmt.Errorf("policy %q is not declared in %s", name, config.ProjectConfigFile)
	}
	cfg.Policies = pruned

	if err := config.WriteProjectConfig(absDir, cfg); err != nil {
		return err
	}

	if err := policies.Reset(name, absDir, charterRoot); err != nil {
		return err
	}

	lf, err := lockfile.Read(absDir, charterRoot)
	if err != nil {
		return err
	}
	delete(lf.Policies, name)
	if err := lockfile.Write(absDir, charterRoot, lf); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "✓ removed %s\n", name)
	return nil
}

func printPolicyRemoveUsage(w *os.File) {
	fmt.Fprint(w, `keystone policy remove — uninstall a policy

Usage:
  keystone policy remove <name> [--dir <path>] [--charter-root <name>]

Removes the named policy from keystone.json, deletes its vendor directory,
and drops its lockfile entry.

Flags:
  --dir <path>           Project root (defaults to cwd).
  --charter-root <name>  Charter directory name.
`)
}
