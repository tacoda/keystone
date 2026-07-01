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

// runInstall handles `keystone install [--dir <path>] [--charter-root <name>]`.
//
// Reads keystone.json from --dir (or cwd), then for each policy in the
// declared tree: fetch the source at its pinned version into the
// content-addressable cache, install the content under <charter-root>/policies/<name>/,
// record per-file hashes in the lockfile. Re-installing existing policies
// rewrites the vendor directory from cache; drift detected post-install is
// reset to the pinned state.
func runInstall(args []string) error {
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
	charterRoot := config.DefaultCharterRoot

	cfg, err := config.ReadProjectConfig(absDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no %s at %s — run `keystone init` first", config.ProjectConfigFile, absDir)
		}
		return err
	}

	if len(cfg.Policies) == 0 {
		fmt.Fprintf(os.Stdout, "✓ %s declares no policies — nothing to install.\n", config.ProjectConfigFile)
		return nil
	}

	lf, err := lockfile.Read(absDir, charterRoot)
	if err != nil {
		return err
	}

	flat := flattenPolicies(cfg.Policies)
	for _, node := range flat {
		if err := installOnePolicy(absDir, charterRoot, node, lf); err != nil {
			return fmt.Errorf("policy %q: %w", node.Name, err)
		}
	}

	if err := lockfile.Write(absDir, charterRoot, lf); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "  wrote: %s\n", filepath.Join(absDir, lockfile.RelPath(charterRoot)))
	fmt.Fprintf(os.Stdout, "✓ installed %d policy(s).\n", len(flat))
	return nil
}

// installOnePolicy runs the fetch → install → record-in-lockfile pipeline
// for a single policy node.
func installOnePolicy(projectDir, charterRoot string, node config.PolicyNode, lf *lockfile.Lockfile) error {
	gitURL := config.ExpandSource(node.Source)
	fmt.Fprintf(os.Stdout, "▸ %s @ %s\n", node.Source, node.Version)

	cached, err := policies.Fetch(gitURL, node.Version)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	installed, err := policies.Install(cached, node.Name, projectDir, charterRoot)
	if err != nil {
		return fmt.Errorf("install: %w", err)
	}

	lf.Policies[node.Name] = lockfile.PolicyLock{
		SourceRef:     node.Source,
		ResolvedSHA:   cached.ResolvedSHA,
		PolicyVersion: installed.PolicyVersion,
		Version:       node.Version,
		Files:         installed.Files,
	}

	fmt.Fprintf(os.Stdout, "  ✓ %d file(s) installed at %s\n",
		len(installed.Files),
		filepath.Join(charterRoot, policies.PolicyRoot, node.Name))
	return nil
}

// flattenPolicies walks the nested policy tree in pre-order and returns a
// flat slice. Mirrors the cascade walk the loader will use in commit 5.
func flattenPolicies(nodes []config.PolicyNode) []config.PolicyNode {
	var out []config.PolicyNode
	for _, n := range nodes {
		out = append(out, n)
		out = append(out, flattenPolicies(n.Children)...)
	}
	return out
}

func printInstallUsage(w *os.File) {
	fmt.Fprint(w, `keystone install — materialize every policy declared in keystone.json

Usage:
  keystone install [--dir <path>] [--charter-root <name>]

Reads keystone.json from <dir> (default: .), then for each policy in the
declared tree: fetches the source at its pinned version into a content-
addressable cache, copies the content into <charter-root>/policies/<name>/,
and records per-file hashes in <charter-root>/keystone.lock.json so drift
can be detected later.

Re-running install is safe: existing vendor directories are rewritten from
the cache, so a stale or hand-edited tree is reset to the pinned state.

Flags:
  --dir <path>           Project root (defaults to cwd).
  --charter-root <name>  Charter directory name (defaults to keystone.json's
                         harness_root, then "charter").
`)
}
