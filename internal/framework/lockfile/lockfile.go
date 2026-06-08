// Package lockfile reads and writes harness/keystone.lock.json, the install-
// state record that pins every installed policy by source ref + resolved SHA
// + per-file hashes. JSON format.
package lockfile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tacoda/keystone/internal/framework/manifest"
)

// File is the on-disk name of the lockfile, relative to the install directory.
// Holds both keystone install state (version, agents, install date) and per-
// policy records.
const File = "harness/keystone.lock.json"

// Version is the schema version of the lockfile format. Bump when fields are
// renamed or removed.
const Version = 1

// KeystoneInfo records the install-scoped state: binary version, install
// date, and agent IDs with menu files installed.
type KeystoneInfo struct {
	Version   string   `json:"version"`
	Installed string   `json:"installed,omitempty"`
	Agents    []string `json:"agents,omitempty"`
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
	Policies map[string]PolicyLock `json:"policies,omitempty"`
}

// Read loads the lockfile from installDir, returning an empty Lockfile (not
// an error) if the file does not exist.
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
	if err := json.Unmarshal(data, &lf); err != nil {
		return nil, fmt.Errorf("parse %s: %w", File, err)
	}
	if lf.Policies == nil {
		lf.Policies = map[string]PolicyLock{}
	}
	return &lf, nil
}

// Write serializes the lockfile back to disk under installDir. Always emits
// Version = lockfile.Version. Indents for human readability.
func Write(installDir string, lf *Lockfile) error {
	lf.Version = Version

	out, err := json.MarshalIndent(lf, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal lockfile: %w", err)
	}
	out = append(out, '\n')

	path := filepath.Join(installDir, File)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, out, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", File, err)
	}
	return nil
}
