package migrations

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Version 2.1 — terminology rename completion. The 2.0 layout move
// renamed the on-disk directory (plugins/ → policies/) and manifest
// filename (keystone-plugin.json → keystone-policy.json). 2.1 finishes
// the rename at the lockfile JSON-tag layer: each policy entry's
// `plugin_version` field becomes `policy_version`.
//
// All other plugin→policy changes shipped in 2.1 are code- or
// documentation-level (package comments, identifiers, env var with a
// backward-compat fallback) and require no on-disk transform.
//
// The lockfile reader keeps a backward-compat fallback for
// `plugin_version` so unmigrated installs still parse cleanly. This
// migration is the active-rewrite path for users who want the lockfile
// on disk to match the current schema.

func init() {
	Register(Migration{
		Version: "2.1",
		Up:      planUp_2_1,
		Down:    planDown_2_1,
	})
}

const v2LockfilePath = ".keystone/lockfile.json"

func planUp_2_1(absDir string) (*Plan, error) {
	p := &Plan{}
	lockPath := filepath.Join(absDir, v2LockfilePath)
	if !fileExists(lockPath) {
		// Nothing to migrate. Fresh install with no lockfile yet — the
		// next install/policy add writes a 2.1-shaped one directly.
		return p, nil
	}
	p.Add("rewrite .keystone/lockfile.json: policies[*].plugin_version → policy_version", func(_ string) error {
		return rewriteLockfilePolicyVersion(lockPath, "plugin_version", "policy_version")
	})
	return p, nil
}

func planDown_2_1(absDir string) (*Plan, error) {
	p := &Plan{}
	lockPath := filepath.Join(absDir, v2LockfilePath)
	if !fileExists(lockPath) {
		return p, nil
	}
	p.Add("rewrite .keystone/lockfile.json: policies[*].policy_version → plugin_version", func(_ string) error {
		return rewriteLockfilePolicyVersion(lockPath, "policy_version", "plugin_version")
	})
	return p, nil
}

// rewriteLockfilePolicyVersion renames a single field on every entry
// under the top-level `policies` object. Pure JSON-level rewrite so the
// transform stays independent of Go's PolicyLock struct (future schema
// changes don't break replay of past migrations).
//
// Idempotent: if no entry holds the from field, nothing is written.
// Conflict (both from AND to present) is left alone — the new field
// already exists, which means the entry was written by a 2.1+ binary
// and the legacy field is stale; don't overwrite real data.
func rewriteLockfilePolicyVersion(path, from, to string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	policiesAny, ok := doc["policies"]
	if !ok {
		return nil
	}
	policies, ok := policiesAny.(map[string]any)
	if !ok {
		return nil
	}
	changed := false
	for name, entryAny := range policies {
		entry, ok := entryAny.(map[string]any)
		if !ok {
			continue
		}
		fromVal, hasFrom := entry[from]
		_, hasTo := entry[to]
		if !hasFrom {
			continue
		}
		if hasTo {
			// New field already present — caller wrote with the new schema.
			// Drop the legacy field rather than overwriting the new one.
			delete(entry, from)
		} else {
			entry[to] = fromVal
			delete(entry, from)
		}
		policies[name] = entry
		changed = true
	}
	if !changed {
		return nil
	}
	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')
	return os.WriteFile(path, out, 0o644)
}
