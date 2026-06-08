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

// runInstall handles `keystone install [--dir <path>] [--harness-root <name>]`.
//
// Reads keystone.json from --dir (or cwd), then for each plugin in the
// declared tree: fetch the source at its pinned version into the
// content-addressable cache, install the content under <harness-root>/plugins/<name>/,
// record per-file hashes in the lockfile. Re-installing existing plugins
// rewrites the vendor directory from cache; drift detected post-install is
// reset to the pinned state.
func runInstall(args []string) error {
	flagValue, args, err := extractHarnessRootFlag(args)
	if err != nil {
		return err
	}
	dir := "."

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--help" || a == "-h":
			printInstallUsage(os.Stdout)
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
			return fmt.Errorf("unexpected positional argument %q", a)
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
			return fmt.Errorf("no %s at %s — run `keystone init` first", config.ProjectConfigFile, absDir)
		}
		return err
	}

	if len(cfg.Plugins) == 0 {
		fmt.Fprintf(os.Stdout, "✓ %s declares no plugins — nothing to install.\n", config.ProjectConfigFile)
		return nil
	}

	lf, err := lockfile.Read(absDir, harnessRoot)
	if err != nil {
		return err
	}

	flat := flattenPlugins(cfg.Plugins)
	for _, node := range flat {
		if err := installOnePlugin(absDir, harnessRoot, node, lf); err != nil {
			return fmt.Errorf("plugin %q: %w", node.Name, err)
		}
	}

	if err := lockfile.Write(absDir, harnessRoot, lf); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "  wrote: %s\n", filepath.Join(absDir, lockfile.RelPath(harnessRoot)))
	fmt.Fprintf(os.Stdout, "✓ installed %d plugin(s).\n", len(flat))
	return nil
}

// installOnePlugin runs the fetch → install → record-in-lockfile pipeline
// for a single plugin node.
func installOnePlugin(projectDir, harnessRoot string, node config.PluginNode, lf *lockfile.Lockfile) error {
	gitURL := config.ExpandSource(node.Source)
	fmt.Fprintf(os.Stdout, "▸ %s @ %s\n", node.Source, node.Version)

	cached, err := plugins.Fetch(gitURL, node.Version)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	installed, err := plugins.Install(cached, node.Name, projectDir, harnessRoot)
	if err != nil {
		return fmt.Errorf("install: %w", err)
	}

	lf.Plugins[node.Name] = lockfile.PluginLock{
		SourceRef:     node.Source,
		ResolvedSHA:   cached.ResolvedSHA,
		PluginVersion: installed.PluginVersion,
		Version:       node.Version,
		Files:         installed.Files,
	}

	fmt.Fprintf(os.Stdout, "  ✓ %d file(s) installed at %s\n",
		len(installed.Files),
		filepath.Join(harnessRoot, plugins.PluginRoot, node.Name))
	return nil
}

// flattenPlugins walks the nested plugin tree in pre-order and returns a
// flat slice. Mirrors the cascade walk the loader will use in commit 5.
func flattenPlugins(nodes []config.PluginNode) []config.PluginNode {
	var out []config.PluginNode
	for _, n := range nodes {
		out = append(out, n)
		out = append(out, flattenPlugins(n.Children)...)
	}
	return out
}

func printInstallUsage(w *os.File) {
	fmt.Fprint(w, `keystone install — materialize every plugin declared in keystone.json

Usage:
  keystone install [--dir <path>] [--harness-root <name>]

Reads keystone.json from <dir> (default: .), then for each plugin in the
declared tree: fetches the source at its pinned version into a content-
addressable cache, copies the content into <harness-root>/plugins/<name>/,
and records per-file hashes in <harness-root>/keystone.lock.json so drift
can be detected later.

Re-running install is safe: existing vendor directories are rewritten from
the cache, so a stale or hand-edited tree is reset to the pinned state.

Flags:
  --dir <path>           Project root (defaults to cwd).
  --harness-root <name>  Harness directory name (defaults to keystone.json's
                         harness_root, then "harness").
`)
}
