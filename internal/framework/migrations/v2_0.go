package migrations

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Version 2.0 — the layout move that lifted harness/ under .keystone/,
// renamed the vendored-policy directory and manifest, and bumped the
// keystone.json schema.
//
// Up steps (each idempotent):
//
//  1. Move harness/ → .keystone/harness/.
//  2. Lift the lockfile from <harnessRoot>/keystone.lock.json to
//     .keystone/lockfile.json.
//  3. Rename .keystone/harness/plugins/ → .keystone/harness/policies/.
//  4. Rename keystone-plugin.json → keystone-policy.json inside every
//     vendored policy.
//  5. Rewrite keystone.json: version "1" → "2", rename top-level
//     `plugins` field to `policies`, strip deprecated `harness_root`.
//
// Down reverses each step in reverse order. harness_root is NOT
// restored on Down (we don't know its prior value); the consumer can
// re-add it by hand if a custom value was needed.

func init() {
	Register(Migration{
		Version: "2.0",
		Up:      planUp_2_0,
		Down:    planDown_2_0,
	})
}

const v2KeystoneJSON = "keystone.json"

func planUp_2_0(absDir string) (*Plan, error) {
	p := &Plan{}

	legacyHarness := filepath.Join(absDir, "harness")
	newHarness := filepath.Join(absDir, ".keystone", "harness")
	keystoneDir := filepath.Join(absDir, ".keystone")

	if dirExists(legacyHarness) && dirExists(newHarness) {
		return nil, fmt.Errorf("both legacy harness/ and .keystone/harness/ exist — resolve by hand before migrating up")
	}

	if dirExists(legacyHarness) {
		p.Add("move harness/ → .keystone/harness/", func(_ string) error {
			if err := os.MkdirAll(keystoneDir, 0o755); err != nil {
				return err
			}
			return os.Rename(legacyHarness, newHarness)
		})
	}

	p.Add("lift lockfile → .keystone/lockfile.json", func(_ string) error {
		newLockPath := filepath.Join(keystoneDir, "lockfile.json")
		for _, cand := range []string{
			filepath.Join(newHarness, "keystone.lock.json"),
			filepath.Join(absDir, "keystone.lock.json"),
		} {
			if fileExists(cand) {
				if err := os.MkdirAll(keystoneDir, 0o755); err != nil {
					return err
				}
				return os.Rename(cand, newLockPath)
			}
		}
		return nil
	})

	pluginsDir := filepath.Join(newHarness, "plugins")
	policiesDir := filepath.Join(newHarness, "policies")
	p.Add("rename .keystone/harness/plugins/ → .keystone/harness/policies/", func(_ string) error {
		if !dirExists(pluginsDir) {
			return nil
		}
		if dirExists(policiesDir) {
			return fmt.Errorf("both plugins/ and policies/ exist in .keystone/harness/ — resolve by hand")
		}
		return os.Rename(pluginsDir, policiesDir)
	})

	p.Add("rename keystone-plugin.json → keystone-policy.json in vendored policies", func(_ string) error {
		if !dirExists(policiesDir) {
			return nil
		}
		return filepath.WalkDir(policiesDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if filepath.Base(path) == "keystone-plugin.json" {
				dst := filepath.Join(filepath.Dir(path), "keystone-policy.json")
				if fileExists(dst) {
					return nil
				}
				return os.Rename(path, dst)
			}
			return nil
		})
	})

	cfgPath := filepath.Join(absDir, v2KeystoneJSON)
	if fileExists(cfgPath) {
		p.Add("rewrite keystone.json: version 2, plugins→policies, strip harness_root", func(_ string) error {
			return rewriteKeystoneJSONUp(cfgPath)
		})
	}

	return p, nil
}

func planDown_2_0(absDir string) (*Plan, error) {
	p := &Plan{}

	legacyHarness := filepath.Join(absDir, "harness")
	newHarness := filepath.Join(absDir, ".keystone", "harness")
	keystoneDir := filepath.Join(absDir, ".keystone")

	if dirExists(legacyHarness) && dirExists(newHarness) {
		return nil, fmt.Errorf("both legacy harness/ and .keystone/harness/ exist — resolve by hand before migrating down")
	}

	cfgPath := filepath.Join(absDir, v2KeystoneJSON)
	if fileExists(cfgPath) {
		p.Add("rewrite keystone.json: version 1, policies→plugins (harness_root not restored)", func(_ string) error {
			return rewriteKeystoneJSONDown(cfgPath)
		})
	}

	pluginsDir := filepath.Join(newHarness, "plugins")
	policiesDir := filepath.Join(newHarness, "policies")
	p.Add("rename keystone-policy.json → keystone-plugin.json in vendored policies", func(_ string) error {
		if !dirExists(policiesDir) {
			return nil
		}
		return filepath.WalkDir(policiesDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if filepath.Base(path) == "keystone-policy.json" {
				dst := filepath.Join(filepath.Dir(path), "keystone-plugin.json")
				if fileExists(dst) {
					return nil
				}
				return os.Rename(path, dst)
			}
			return nil
		})
	})

	p.Add("rename .keystone/harness/policies/ → .keystone/harness/plugins/", func(_ string) error {
		if !dirExists(policiesDir) {
			return nil
		}
		if dirExists(pluginsDir) {
			return fmt.Errorf("both plugins/ and policies/ exist in .keystone/harness/ — resolve by hand")
		}
		return os.Rename(policiesDir, pluginsDir)
	})

	p.Add("drop lockfile back to <harnessRoot>/keystone.lock.json", func(_ string) error {
		newLockPath := filepath.Join(keystoneDir, "lockfile.json")
		if !fileExists(newLockPath) {
			return nil
		}
		// Restore to the pre-2.0 location inside the harness tree.
		// The harness is still under .keystone/harness/ at this point;
		// the harness-move step below relocates it back. We drop the
		// lockfile inside the harness so the harness-rename carries it.
		destDir := newHarness
		if !dirExists(destDir) {
			// Harness already moved back — drop next to it instead.
			destDir = legacyHarness
			if !dirExists(destDir) {
				if err := os.MkdirAll(destDir, 0o755); err != nil {
					return err
				}
			}
		}
		return os.Rename(newLockPath, filepath.Join(destDir, "keystone.lock.json"))
	})

	if dirExists(newHarness) {
		p.Add("move .keystone/harness/ → harness/", func(_ string) error {
			if err := os.Rename(newHarness, legacyHarness); err != nil {
				return err
			}
			// Best-effort: remove the now-empty .keystone/ dir.
			_ = os.Remove(keystoneDir)
			return nil
		})
	}

	return p, nil
}

// rewriteKeystoneJSONUp applies 1.x → 2.0 schema transforms in place.
// Idempotent: re-running on already-2.0 content is a no-op.
func rewriteKeystoneJSONUp(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return fmt.Errorf("parse keystone.json: %w", err)
	}
	changed := false
	if v, ok := doc["version"]; ok {
		if s, _ := v.(string); s != "2" {
			doc["version"] = "2"
			changed = true
		}
	} else {
		doc["version"] = "2"
		changed = true
	}
	if _, ok := doc["plugins"]; ok {
		if _, alreadyPolicies := doc["policies"]; alreadyPolicies {
			delete(doc, "plugins")
		} else {
			doc["policies"] = doc["plugins"]
			delete(doc, "plugins")
		}
		changed = true
	}
	if _, ok := doc["harness_root"]; ok {
		delete(doc, "harness_root")
		changed = true
	}
	if !changed {
		return nil
	}
	return writeOrderedJSON(path, doc)
}

// rewriteKeystoneJSONDown reverses the up transforms. Idempotent.
func rewriteKeystoneJSONDown(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return fmt.Errorf("parse keystone.json: %w", err)
	}
	changed := false
	if v, ok := doc["version"]; ok {
		if s, _ := v.(string); s != "1" {
			doc["version"] = "1"
			changed = true
		}
	}
	if _, ok := doc["policies"]; ok {
		if _, alreadyPlugins := doc["plugins"]; alreadyPlugins {
			delete(doc, "policies")
		} else {
			doc["plugins"] = doc["policies"]
			delete(doc, "policies")
		}
		changed = true
	}
	if !changed {
		return nil
	}
	return writeOrderedJSON(path, doc)
}

func writeOrderedJSON(path string, doc map[string]any) error {
	ordered := map[string]any{}
	for _, k := range []string{"version", "framework_version", "policies", "plugins", "budgets"} {
		if v, ok := doc[k]; ok {
			ordered[k] = v
			delete(doc, k)
		}
	}
	for k, v := range doc {
		ordered[k] = v
	}
	out, err := json.MarshalIndent(ordered, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')
	return os.WriteFile(path, out, 0o644)
}

func dirExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}
