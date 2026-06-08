// Package lockfile reads and writes harness/.keystone.lock, the install-state
// record that pins every installed policy by source ref + resolved SHA + per-
// file hashes. The format is YAML at 0.x and JSON at 1.0 (Phase 1 commit 5).
// The Go types are stable across that format switch.
package lockfile

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/tacoda/keystone/internal/framework/manifest"
)

// File is the on-disk name of the combined lockfile, relative to the install
// directory. Holds both keystone install state (version, agents, install
// date) and per-policy records.
const File = "harness/.keystone.lock"

// Version is the schema version of the lockfile format. Bump when fields are
// renamed or removed.
const Version = 1

// KeystoneInfo records the install-scoped state that used to live in
// INSTALL_PROFILE.md frontmatter (version) and the agent row (agents).
// Authoritative once the lockfile exists; INSTALL_PROFILE.md remains
// human-readable but is no longer parsed by machine code.
type KeystoneInfo struct {
	Version   string   `yaml:"version"`             // binary version that last touched the install
	Installed string   `yaml:"installed,omitempty"` // YYYY-MM-DD of original install
	Agents    []string `yaml:"agents,omitempty"`    // agent IDs with menu files installed
}

// PolicyLock describes one installed policy. Tier, Strict, and Required are
// recorded at install time so `keystone policy verify` can run without re-
// resolving the source policy.
type PolicyLock struct {
	SourceRef       string              `yaml:"source_ref"`         // exact ref string the user passed
	ResolvedSHA     string              `yaml:"resolved_sha"`       // commit SHA / artifact digest
	PolicyVersion   string              `yaml:"policy_version"`     // value from manifest.version
	KeystoneVersion string              `yaml:"keystone_version"`   // binary version at install time
	Tier            string              `yaml:"tier,omitempty"`     // "org" (default) or "team"
	Strict          manifest.StrictSpec `yaml:"strict,omitempty"`   // items this policy locks against override
	Required        manifest.StrictSpec `yaml:"required,omitempty"` // items this policy expects to exist somewhere in the cascade
	Files           map[string]string   `yaml:"files"`              // path-relative-to-installdir → "sha256:<hex>"
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
	Version  int                   `yaml:"version"`
	Keystone KeystoneInfo          `yaml:"keystone"`
	Policies map[string]PolicyLock `yaml:"policies,omitempty"`
}

// Read loads the lockfile from installDir, returning an empty Lockfile
// (not an error) if the file does not exist.
func Read(installDir string) (*Lockfile, error) {
	path := filepath.Join(installDir, File)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Lockfile{Version: Version, Policies: map[string]PolicyLock{}}, nil
		}
		return nil, fmt.Errorf("read %s: %w", File, err)
	}
	var lf Lockfile
	if err := yaml.Unmarshal(data, &lf); err != nil {
		return nil, fmt.Errorf("parse %s: %w", File, err)
	}
	if lf.Policies == nil {
		lf.Policies = map[string]PolicyLock{}
	}
	return &lf, nil
}

// Write serializes the lockfile back to disk under installDir. Always emits
// Version = lockfile.Version.
func Write(installDir string, lf *Lockfile) error {
	lf.Version = Version

	out, err := yaml.Marshal(lf)
	if err != nil {
		return fmt.Errorf("marshal lockfile: %w", err)
	}

	path := filepath.Join(installDir, File)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, out, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", File, err)
	}
	return nil
}
