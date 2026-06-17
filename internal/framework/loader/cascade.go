package loader

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/manifest"
	"github.com/tacoda/keystone/internal/framework/policies"
)

// VerifyResult is the outcome of a cascade verification: per-plugin drift
// reports, per-port strict-cascade violations, depth-rule violations
// (sensors at deeper-than-top plugins), and advisory required-item gaps.
// Drift and required gaps are informational; strict and depth violations
// are hard errors that block clean installs.
type VerifyResult struct {
	Violations      []ShadowViolation
	DepthViolations []DepthViolation
	RequiredGaps    []RequiredGap
	Drift           []PolicyDrift
}

// HasErrors reports whether any strict or depth rule was violated. Drift
// and required gaps are advisory.
func (r VerifyResult) HasErrors() bool {
	return len(r.Violations) > 0 || len(r.DepthViolations) > 0
}

// HasGaps reports whether any plugin's `required` item was not satisfied.
// Advisory only — does not affect HasErrors.
func (r VerifyResult) HasGaps() bool { return len(r.RequiredGaps) > 0 }

// HasDrift reports whether any vendored plugin diverges from its
// lockfile entry.
func (r VerifyResult) HasDrift() bool { return len(r.Drift) > 0 }

// ShadowViolation reports a strict item being overridden by a deeper node
// in the cascade. PathContext renders the tree-path (e.g.
// "acme-org > acme-platform") to the offending node.
type ShadowViolation struct {
	Policy      string
	PathContext string
	Port        string
	Item        string
	ShadowPaths []string
}

func (v ShadowViolation) String() string {
	return fmt.Sprintf("plugin %q (%s) marks %s/%s strict — overridden by:\n    %s",
		v.Policy, v.PathContext, v.Port, v.Item, strings.Join(v.ShadowPaths, "\n    "))
}

// DepthViolation reports a nested plugin that ships or strict-claims
// sensors. Sensors are only allowed at the project layer and at top-level
// plugins in the consumer's keystone.json; plugins nested below another
// plugin may not contribute sensors.
type DepthViolation struct {
	Policy          string
	PathContext     string
	Depth           int      // number of plugin ancestors above this node
	StrictSensors   []string // sensor names this nested node strict-claims
	VendoredSensors []string // sensor files this nested node ships
}

func (v DepthViolation) String() string {
	parts := []string{}
	if len(v.StrictSensors) > 0 {
		parts = append(parts, fmt.Sprintf("strict.sensors: %s", strings.Join(v.StrictSensors, ", ")))
	}
	if len(v.VendoredSensors) > 0 {
		parts = append(parts, fmt.Sprintf("shipped sensors:\n    %s", strings.Join(v.VendoredSensors, "\n    ")))
	}
	return fmt.Sprintf("plugin %q (%s) is nested at depth %d — sensors are only allowed at the project layer and at top-level plugins:\n  %s",
		v.Policy, v.PathContext, v.Depth, strings.Join(parts, "\n  "))
}

// RequiredGap reports a plugin that declares a `required` item that no
// outer layer (any plugin shallower in keystone.json, or the project)
// provides. The gap is advisory — `keystone verify` surfaces it but does
// not fail; solo installs may legitimately defer providing the item.
type RequiredGap struct {
	Policy      string
	PathContext string
	Port        string
	Item        string
}

func (g RequiredGap) String() string {
	return fmt.Sprintf("plugin %q (%s) requires %s/%s — define it at <harness-root>/%s/%s.md (or in an outer plugin)",
		g.Policy, g.PathContext, g.Port, g.Item, g.Port, g.Item)
}

// PolicyDrift reports a single drifted plugin: which files differ from the
// lockfile, and what kind of difference.
type PolicyDrift struct {
	Policy string
	Files  []policies.Drift
}

// Verify walks the plugin tree from keystone.json, checks each plugin's
// vendor directory against the lockfile's per-file hashes, and reports
// strict-cascade violations.
//
// Drift detection is a precondition for clean cascade resolution: the
// runtime should reset any drifted plugin via policies.Reset before
// reading its content. This function returns drift; it does not perform
// the reset.
//
// Strict semantics: a plugin's strict map names items it locks absolutely.
// Nothing else can override a strict item — not the project, not any other
// plugin (sibling, ancestor, or descendant). This walker currently catches
// the most common case: project files shadowing a strict item. Policy-on-
// plugin strict conflicts are refused at install time and detected at
// load time by file count.
//
// Default (non-strict) cascade resolution: the project always wins; among
// plugins, deeper-nested plugins (children in keystone.json) refine the
// outer plugins they're nested in.
func Verify(installDir string, cfg *config.ProjectConfig, expectedFiles map[string]map[string]string) (*VerifyResult, error) {
	harnessRoot := cfg.ResolvedHarnessRoot()
	res := &VerifyResult{}

	// Walk pre-order: each node carries its breadcrumb path for violation
	// messages.
	visit := func(node config.PolicyNode, path []string) error {
		ctx := strings.Join(append(path, node.Name), " > ")
		exp := expectedFiles[node.Name]
		drifts, err := policies.Verify(node.Name, installDir, harnessRoot, exp)
		if err != nil {
			return fmt.Errorf("verify plugin %q: %w", node.Name, err)
		}
		if len(drifts) > 0 {
			res.Drift = append(res.Drift, PolicyDrift{Policy: node.Name, Files: drifts})
		}
		// Depth gate: nested plugins (any plugin with a plugin ancestor) may
		// not contribute sensors. Top-level plugins and the project layer can.
		if len(path) > 0 {
			depthV := DepthViolation{
				Policy:        node.Name,
				PathContext:   ctx,
				Depth:         len(path),
				StrictSensors: append([]string{}, node.Strict["sensors"]...),
			}
			sensorsPrefix := filepath.ToSlash(filepath.Join(harnessRoot, "policies", node.Name, "sensors")) + "/"
			for rel := range exp {
				if strings.HasPrefix(rel, sensorsPrefix) {
					depthV.VendoredSensors = append(depthV.VendoredSensors, rel)
				}
			}
			if len(depthV.StrictSensors) > 0 || len(depthV.VendoredSensors) > 0 {
				res.DepthViolations = append(res.DepthViolations, depthV)
			}
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
					Policy:      node.Name,
					PathContext: ctx,
					Port:        port,
					Item:        item,
					ShadowPaths: shadowed,
				})
			}
		}
		// Required-item gaps: each item this plugin declares as `required`
		// must be satisfied by an outer layer (an ancestor in the plugin
		// tree, or the project). Siblings and descendants do NOT satisfy
		// required. Missing items are advisory gaps, not errors.
		m, err := loadInstalledManifest(installDir, harnessRoot, node.Name)
		if err != nil {
			return fmt.Errorf("load manifest for %q: %w", node.Name, err)
		}
		if m != nil {
			for port, items := range requiredByPort(m.Required) {
				for _, item := range items {
					if requiredSatisfied(installDir, harnessRoot, path, port, item) {
						continue
					}
					res.RequiredGaps = append(res.RequiredGaps, RequiredGap{
						Policy:      node.Name,
						PathContext: ctx,
						Port:        port,
						Item:        item,
					})
				}
			}
		}
		return nil
	}

	var walk func(nodes []config.PolicyNode, path []string) error
	walk = func(nodes []config.PolicyNode, path []string) error {
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
	if err := walk(cfg.Policies, nil); err != nil {
		return nil, err
	}
	return res, nil
}

// findShadowing returns the project-layer paths (under <harnessRoot>/<port>/)
// that override a strict item declared by `plugin`. Strict is absolute, so
// any project file with the matching basename is a violation regardless of
// which plugin declared the strict.
//
// Policy-on-plugin strict shadowing is not surfaced here: at 1.0 plugins are
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

// loadInstalledManifest reads the vendored keystone-plugin.json for `plugin`
// from <installDir>/<harnessRoot>/plugins/<plugin>/. Returns nil with no
// error when the manifest is missing (older installs may omit it); the
// caller should treat absence as "no required claims to check."
func loadInstalledManifest(installDir, harnessRoot, plugin string) (*manifest.Manifest, error) {
	pluginRoot := filepath.Join(installDir, harnessRoot, "policies", plugin)
	if _, err := os.Stat(filepath.Join(pluginRoot, manifest.PolicyManifestFile)); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return manifest.Load(pluginRoot)
}

// requiredByPort flattens a StrictSpec into a port-keyed map (mirroring
// PolicyNode.Strict's shape) so the verify walker can iterate uniformly.
func requiredByPort(s manifest.StrictSpec) map[string][]string {
	out := map[string][]string{}
	if len(s.Guides) > 0 {
		out["guides"] = s.Guides
	}
	if len(s.Playbooks) > 0 {
		out["playbooks"] = s.Playbooks
	}
	if len(s.Actions) > 0 {
		out["actions"] = s.Actions
	}
	if len(s.Sensors) > 0 {
		out["sensors"] = s.Sensors
	}
	return out
}

// requiredSatisfied reports whether an item required by a plugin at depth
// len(path) is supplied by an outer layer — either by a plugin ancestor in
// the same path, or by the project layer. Siblings and descendants do not
// satisfy `required`.
func requiredSatisfied(installDir, harnessRoot string, ancestors []string, port, item string) bool {
	want := item + ".md"
	// Project layer.
	projectPath := filepath.Join(installDir, harnessRoot, port, want)
	if _, err := os.Stat(projectPath); err == nil {
		return true
	}
	// Ancestor plugins, in order outer → inner.
	for _, anc := range ancestors {
		ancPath := filepath.Join(installDir, harnessRoot, "policies", anc, port, want)
		if _, err := os.Stat(ancPath); err == nil {
			return true
		}
	}
	return false
}
