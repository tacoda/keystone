package loader

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/plugins"
)

// VerifyResult is the outcome of a cascade verification: per-plugin drift
// reports plus per-port strict-cascade violations. Drift is informational
// (the runtime will reset on next load); violations are hard errors that
// block clean installs.
type VerifyResult struct {
	Violations []ShadowViolation
	Drift      []PluginDrift
}

// HasErrors reports whether any strict rule was violated. Drift is
// advisory.
func (r VerifyResult) HasErrors() bool { return len(r.Violations) > 0 }

// HasDrift reports whether any vendored plugin diverges from its
// lockfile entry.
func (r VerifyResult) HasDrift() bool { return len(r.Drift) > 0 }

// ShadowViolation reports a strict item being overridden by a deeper node
// in the cascade. PathContext renders the tree-path (e.g.
// "acme-org > acme-platform") to the offending node.
type ShadowViolation struct {
	Plugin      string
	PathContext string
	Port        string
	Item        string
	ShadowPaths []string
}

func (v ShadowViolation) String() string {
	return fmt.Sprintf("plugin %q (%s) marks %s/%s strict — overridden by:\n    %s",
		v.Plugin, v.PathContext, v.Port, v.Item, strings.Join(v.ShadowPaths, "\n    "))
}

// PluginDrift reports a single drifted plugin: which files differ from the
// lockfile, and what kind of difference.
type PluginDrift struct {
	Plugin string
	Files  []plugins.Drift
}

// Verify walks the plugin tree from keystone.json, checks each plugin's
// vendor directory against the lockfile's per-file hashes, and reports
// strict-cascade violations.
//
// Drift detection is a precondition for clean cascade resolution: the
// runtime should reset any drifted plugin via plugins.Reset before
// reading its content. This function returns drift; it does not perform
// the reset.
//
// Strict semantics: a plugin's strict map names items it locks against
// override from any deeper node in the tree (descendants + later-tree
// siblings whose path depth exceeds this one). The project layer (the
// harness root itself) is treated as a deeper-than-anything node, so any
// project file shadowing a strict item is a violation regardless of where
// the strict declaration lives.
func Verify(installDir string, cfg *config.ProjectConfig, expectedFiles map[string]map[string]string) (*VerifyResult, error) {
	harnessRoot := cfg.ResolvedHarnessRoot()
	res := &VerifyResult{}

	// Walk pre-order: each node carries its breadcrumb path for violation
	// messages.
	visit := func(node config.PluginNode, path []string) error {
		ctx := strings.Join(append(path, node.Name), " > ")
		exp := expectedFiles[node.Name]
		drifts, err := plugins.Verify(node.Name, installDir, harnessRoot, exp)
		if err != nil {
			return fmt.Errorf("verify plugin %q: %w", node.Name, err)
		}
		if len(drifts) > 0 {
			res.Drift = append(res.Drift, PluginDrift{Plugin: node.Name, Files: drifts})
		}
		for port, items := range node.Strict {
			for _, item := range items {
				shadowed, err := findShadowing(installDir, harnessRoot, node.Name, port, item)
				if err != nil {
					return err
				}
				if len(shadowed) == 0 {
					continue
				}
				res.Violations = append(res.Violations, ShadowViolation{
					Plugin:      node.Name,
					PathContext: ctx,
					Port:        port,
					Item:        item,
					ShadowPaths: shadowed,
				})
			}
		}
		return nil
	}

	var walk func(nodes []config.PluginNode, path []string) error
	walk = func(nodes []config.PluginNode, path []string) error {
		for _, n := range nodes {
			if err := visit(n, path); err != nil {
				return err
			}
			if err := walk(n.Children, append(path, n.Name)); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(cfg.Plugins, nil); err != nil {
		return nil, err
	}
	return res, nil
}

// findShadowing returns the project-layer paths (under <harnessRoot>/<port>/)
// that override a strict item declared by `plugin`. The project layer is
// always treated as deeper than any plugin node, so any project file with
// the matching basename counts.
//
// Sibling-plugin shadowing is not surfaced here: at 1.0 plugins are
// vendored read-only, and `keystone install` would refuse to write a
// strict-violating file in the first place. The check is for the
// project layer's free-form content.
func findShadowing(installDir, harnessRoot, plugin, port, item string) ([]string, error) {
	root := filepath.Join(installDir, harnessRoot, port)
	want := item + ".md"

	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil, nil
	}

	var hits []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Base(path) != want {
			return nil
		}
		rel, relErr := filepath.Rel(installDir, path)
		if relErr != nil {
			return relErr
		}
		hits = append(hits, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return hits, nil
}
