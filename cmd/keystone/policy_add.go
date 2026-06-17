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

// runPolicyAdd handles `keystone policy add <shorthand> [--name <n>] [--dir <path>] [--harness-root <name>]`.
//
// Parses the shorthand into source+version, derives a default name (the
// last source path segment) unless --name overrides, appends a PolicyNode
// to keystone.json, then runs the same fetch + install + lockfile-record
// pipeline that `keystone install` uses.
func runPolicyAdd(args []string) error {
	dir := "."
	var nameOverride string
	var positional []string

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--help" || a == "-h":
			printPolicyAddUsage(os.Stdout)
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
		return fmt.Errorf("policy add requires exactly one shorthand argument (e.g. `keystone policy add tacoda/tacoda-org@0.2.0`)")
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
		name = config.DefaultPolicyName(source)
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	harnessRoot := config.DefaultHarnessRoot

	cfg, err := loadOrCreateProjectConfig(absDir, harnessRoot)
	if err != nil {
		return err
	}

	if findPolicy(cfg.Policies, name) != nil {
		return fmt.Errorf("policy %q is already declared in %s — use `keystone policy update` to change its version", name, config.ProjectConfigFile)
	}

	node := config.PolicyNode{
		Name:    name,
		Source:  source,
		Version: version,
	}
	cfg.Policies = append(cfg.Policies, node)

	if err := config.WriteProjectConfig(absDir, cfg); err != nil {
		return err
	}

	lf, err := lockfile.Read(absDir, harnessRoot)
	if err != nil {
		return err
	}
	if err := installOnePolicy(absDir, harnessRoot, node, lf); err != nil {
		return err
	}
	if err := lockfile.Write(absDir, harnessRoot, lf); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "✓ added %s @ %s as %q\n", source, version, name)
	return nil
}

// loadOrCreateProjectConfig returns the parsed keystone.json, or a
// freshly-defaulted one if no file exists yet. This makes `policy add` work
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

// findPolicy walks the nested tree and returns a pointer to the matching
// node (or nil) so callers can mutate it in place.
func findPolicy(nodes []config.PolicyNode, name string) *config.PolicyNode {
	for i := range nodes {
		if nodes[i].Name == name {
			return &nodes[i]
		}
		if hit := findPolicy(nodes[i].Children, name); hit != nil {
			return hit
		}
	}
	return nil
}

// removePolicy walks the nested tree and returns a new slice with the
// named node removed. Returns (slice, true) on hit, (slice, false) on miss.
func removePolicy(nodes []config.PolicyNode, name string) ([]config.PolicyNode, bool) {
	out := make([]config.PolicyNode, 0, len(nodes))
	removed := false
	for _, n := range nodes {
		if n.Name == name {
			removed = true
			continue
		}
		if pruned, hit := removePolicy(n.Children, name); hit {
			n.Children = pruned
			removed = true
		}
		out = append(out, n)
	}
	return out, removed
}

func printPolicyAddUsage(w *os.File) {
	fmt.Fprint(w, `keystone policy add — install a new policy

Usage:
  keystone policy add <shorthand> [--name <n>] [--dir <path>] [--harness-root <name>]

Shorthand: [<host>/]<owner>/<repo>@<version>. The host defaults to
github.com when omitted; the version is required and is a git ref (tag,
branch, or SHA).

Examples:
  keystone policy add tacoda/tacoda-org@0.2.0
  keystone policy add gitlab.com/acme/policies@v1.0
  keystone policy add github.com/acme/policies@main --name acme

The policy is appended at the top level of keystone.json's policies array.
To nest under another policy, edit keystone.json by hand and re-run
`+"`keystone install`"+`.

Flags:
  --name <n>             Override the default policy name (last source segment).
  --dir <path>           Project root (defaults to cwd).
  --harness-root <name>  Harness directory name (defaults to keystone.json's
                         harness_root, then "harness").
`)
}
