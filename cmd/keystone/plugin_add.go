package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/lockfile"
)

// runPluginAdd handles `keystone plugin add <shorthand> [--name <n>] [--dir <path>] [--harness-root <name>]`.
//
// Parses the shorthand into source+version, derives a default name (the
// last source path segment) unless --name overrides, appends a PluginNode
// to keystone.json, then runs the same fetch + install + lockfile-record
// pipeline that `keystone install` uses.
func runPluginAdd(args []string) error {
	flagValue, args, err := extractHarnessRootFlag(args)
	if err != nil {
		return err
	}
	dir := "."
	var nameOverride string
	var positional []string

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--help" || a == "-h":
			printPluginAddUsage(os.Stdout)
			return nil
		case a == "--dir":
			if i+1 >= len(args) {
				return fmt.Errorf("flag %s requires a value", a)
			}
			dir = args[i+1]
			i++
		case strings.HasPrefix(a, "--dir="):
			dir = strings.TrimPrefix(a, "--dir=")
		case a == "--name":
			if i+1 >= len(args) {
				return fmt.Errorf("flag %s requires a value", a)
			}
			nameOverride = args[i+1]
			i++
		case strings.HasPrefix(a, "--name="):
			nameOverride = strings.TrimPrefix(a, "--name=")
		case strings.HasPrefix(a, "-"):
			return fmt.Errorf("unknown flag %s", a)
		default:
			positional = append(positional, a)
		}
	}

	if len(positional) != 1 {
		return fmt.Errorf("plugin add requires exactly one shorthand argument (e.g. `keystone plugin add tacoda/tacoda-org@0.2.0`)")
	}
	spec := positional[0]
	source, version := config.ParseShorthand(spec)
	if source == "" || version == "" {
		return fmt.Errorf("shorthand %q must be `[<host>/]<owner>/<repo>@<version>`", spec)
	}
	if err := config.ValidateSource(source); err != nil {
		return err
	}

	name := nameOverride
	if name == "" {
		name = config.DefaultPluginName(source)
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	harnessRoot, err := resolveHarnessRoot(absDir, flagValue)
	if err != nil {
		return err
	}

	cfg, err := loadOrCreateProjectConfig(absDir, harnessRoot)
	if err != nil {
		return err
	}

	if findPlugin(cfg.Plugins, name) != nil {
		return fmt.Errorf("plugin %q is already declared in %s — use `keystone plugin update` to change its version", name, config.ProjectConfigFile)
	}

	node := config.PluginNode{
		Name:    name,
		Source:  source,
		Version: version,
	}
	cfg.Plugins = append(cfg.Plugins, node)

	if err := config.WriteProjectConfig(absDir, cfg); err != nil {
		return err
	}

	lf, err := lockfile.Read(absDir, harnessRoot)
	if err != nil {
		return err
	}
	if err := installOnePlugin(absDir, harnessRoot, node, lf); err != nil {
		return err
	}
	if err := lockfile.Write(absDir, harnessRoot, lf); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "✓ added %s @ %s as %q\n", source, version, name)
	return nil
}

// loadOrCreateProjectConfig returns the parsed keystone.json, or a
// freshly-defaulted one if no file exists yet. This makes `plugin add` work
// before `init` has written a keystone.json — the next write persists the
// new file.
func loadOrCreateProjectConfig(projectDir, harnessRoot string) (*config.ProjectConfig, error) {
	cfg, err := config.ReadProjectConfig(projectDir)
	if err == nil {
		return cfg, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return config.DefaultProjectConfig(harnessRoot), nil
	}
	return nil, err
}

// findPlugin walks the nested tree and returns a pointer to the matching
// node (or nil) so callers can mutate it in place.
func findPlugin(nodes []config.PluginNode, name string) *config.PluginNode {
	for i := range nodes {
		if nodes[i].Name == name {
			return &nodes[i]
		}
		if hit := findPlugin(nodes[i].Children, name); hit != nil {
			return hit
		}
	}
	return nil
}

// removePlugin walks the nested tree and returns a new slice with the
// named node removed. Returns (slice, true) on hit, (slice, false) on miss.
func removePlugin(nodes []config.PluginNode, name string) ([]config.PluginNode, bool) {
	out := make([]config.PluginNode, 0, len(nodes))
	removed := false
	for _, n := range nodes {
		if n.Name == name {
			removed = true
			continue
		}
		if pruned, hit := removePlugin(n.Children, name); hit {
			n.Children = pruned
			removed = true
		}
		out = append(out, n)
	}
	return out, removed
}

func printPluginAddUsage(w *os.File) {
	fmt.Fprint(w, `keystone plugin add — install a new plugin

Usage:
  keystone plugin add <shorthand> [--name <n>] [--dir <path>] [--harness-root <name>]

Shorthand: [<host>/]<owner>/<repo>@<version>. The host defaults to
github.com when omitted; the version is required and is a git ref (tag,
branch, or SHA).

Examples:
  keystone plugin add tacoda/tacoda-org@0.2.0
  keystone plugin add gitlab.com/acme/policies@v1.0
  keystone plugin add github.com/acme/policies@main --name acme

The plugin is appended at the top level of keystone.json's plugins array.
To nest under another plugin, edit keystone.json by hand and re-run
`+"`keystone install`"+`.

Flags:
  --name <n>             Override the default plugin name (last source segment).
  --dir <path>           Project root (defaults to cwd).
  --harness-root <name>  Harness directory name (defaults to keystone.json's
                         harness_root, then "harness").
`)
}
