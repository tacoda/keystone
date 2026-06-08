// Package lockfile reads and writes the per-install state record at
// <harness-root>/keystone.lock.json. Pins every installed policy by source
// ref + resolved SHA + per-file hashes. JSON format.
//
// The harness root is configurable (default "harness", overridable per
// install via the --harness-root flag at `keystone init`), so every
// function accepts it explicitly.
package lockfile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/manifest"
)

// Version is the schema version of the lockfile format. Bump when fields are
// renamed or removed.
const Version = 1

// RelPath returns the lockfile's path relative to the install directory:
// <harness-root>/keystone.lock.json. Pass to filepath.Join(installDir, ...)
// for the absolute path.
func RelPath(harnessRoot string) string {
	return filepath.Join(harnessRoot, config.LockfileName)
}

// KeystoneInfo records the install-scoped state: binary version, install
// date, and agent IDs with menu files installed. The harness root used to
// live here too as a transitional Phase 2 affordance; at 1.0 keystone.json
// owns that field instead, so commands resolve it from there.
type KeystoneInfo struct {
	Version   string   `json:"version"`
	Installed string   `json:"installed,omitempty"`
	Agents    []string `json:"agents,omitempty"`
}

// PluginLock describes one installed plugin: where it came from, the exact
// commit it resolved to, and per-file content hashes used by the drift
// detector. Written by `keystone install` / `keystone plugin add|update`,
// consumed by `keystone verify` and the loader's drift-reset path.
type PluginLock struct {
	SourceRef     string            `json:"source_ref"`     // shorthand string from keystone.json (e.g. "tacoda/tacoda-org")
	ResolvedSHA   string            `json:"resolved_sha"`   // exact commit hash resolved during fetch
	PluginVersion string            `json:"plugin_version"` // value from the plugin's manifest, if available
	Version       string            `json:"version"`        // ref the consumer pinned (tag, branch, SHA)
	Files         map[string]string `json:"files"`          // path-relative-to-installdir → "sha256:<hex>"
}

// PolicyLock describes one installed policy. Tier, Strict, and Required are
// recorded at install time so `keystone policy verify` can run without re-
// resolving the source policy.
type PolicyLock struct {
	SourceRef       string              `json:"source_ref"`
	ResolvedSHA     string              `json:"resolved_sha"`
	PolicyVersion   string              `json:"policy_version"`
	KeystoneVersion string              `json:"keystone_version"`
	Tier            string              `json:"tier,omitempty"`
	Strict          manifest.StrictSpec `json:"strict,omitempty"`
	Required        manifest.StrictSpec `json:"required,omitempty"`
	Files           map[string]string   `json:"files"`
}

// ResolvedTier returns the lock's tier, applying the default for older
// lockfiles that predate the field.
func (p PolicyLock) ResolvedTier() string {
	if p.Tier == "" {
		return manifest.TierOrg
	}
	return p.Tier
}

// Lockfile is the root document.
type Lockfile struct {
	Version  int                   `json:"version"`
	Keystone KeystoneInfo          `json:"keystone"`
	Plugins  map[string]PluginLock `json:"plugins,omitempty"`
	Policies map[string]PolicyLock `json:"policies,omitempty"`
}

// Read loads the lockfile at <installDir>/<harnessRoot>/keystone.lock.json,
// returning an empty Lockfile (not an error) if the file does not exist.
func Read(installDir, harnessRoot string) (*Lockfile, error) {
	rel := RelPath(harnessRoot)
	path := filepath.Join(installDir, rel)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Lockfile{
				Version:  Version,
				Plugins:  map[string]PluginLock{},
				Policies: map[string]PolicyLock{},
			}, nil
		}
		return nil, fmt.Errorf("read %s: %w", rel, err)
	}
	var lf Lockfile
	if err := json.Unmarshal(data, &lf); err != nil {
		return nil, fmt.Errorf("parse %s: %w", rel, err)
	}
	if lf.Policies == nil {
		lf.Policies = map[string]PolicyLock{}
	}
	if lf.Plugins == nil {
		lf.Plugins = map[string]PluginLock{}
	}
	return &lf, nil
}

// Write serializes the lockfile back to disk at
// <installDir>/<harnessRoot>/keystone.lock.json. Always emits Version =
// lockfile.Version. Indents for human readability.
func Write(installDir, harnessRoot string, lf *Lockfile) error {
	lf.Version = Version

	out, err := json.MarshalIndent(lf, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal lockfile: %w", err)
	}
	out = append(out, '\n')

	rel := RelPath(harnessRoot)
	path := filepath.Join(installDir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, out, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", rel, err)
	}
	return nil
}
