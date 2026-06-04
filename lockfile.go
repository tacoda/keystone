package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// KeystoneLockfile is the on-disk name of the combined lockfile, relative to
// the install directory. Holds both keystone install state (version, agents,
// install date) and per-policy records.
const KeystoneLockfile = "harness/.keystone.lock"

// LockfileVersion is the schema version of the lockfile format. Bump when
// fields are renamed or removed.
const LockfileVersion = 1

// KeystoneInfo records the install-scoped state that used to live in
// INSTALL_PROFILE.md frontmatter (version) and the agent row (agents).
// Authoritative once the lockfile exists; INSTALL_PROFILE.md remains
// human-readable but is no longer parsed by machine code.
type KeystoneInfo struct {
	Version   string   `yaml:"version"`             // binary version that last touched the install
	Installed string   `yaml:"installed,omitempty"` // YYYY-MM-DD of original install
	Agents    []string `yaml:"agents,omitempty"`    // agent IDs with menu files installed
}

// PolicyLock describes one installed policy (org pack).
type PolicyLock struct {
	SourceRef       string            `yaml:"source_ref"`       // exact ref string the user passed
	ResolvedSHA     string            `yaml:"resolved_sha"`     // commit SHA / artifact digest
	PolicyVersion   string            `yaml:"policy_version"`   // value from manifest.version
	KeystoneVersion string            `yaml:"keystone_version"` // binary version at install time
	Files           map[string]string `yaml:"files"`            // path-relative-to-installdir → "sha256:<hex>"
}

// Lockfile is the root document.
type Lockfile struct {
	Version  int                   `yaml:"version"`
	Keystone KeystoneInfo          `yaml:"keystone"`
	Policies map[string]PolicyLock `yaml:"policies,omitempty"`
}

// readLockfile loads the lockfile from installDir, returning an empty
// Lockfile (not an error) if the file does not exist.
func readLockfile(installDir string) (*Lockfile, error) {
	path := filepath.Join(installDir, KeystoneLockfile)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Lockfile{Version: LockfileVersion, Policies: map[string]PolicyLock{}}, nil
		}
		return nil, fmt.Errorf("read %s: %w", KeystoneLockfile, err)
	}
	var lf Lockfile
	if err := yaml.Unmarshal(data, &lf); err != nil {
		return nil, fmt.Errorf("parse %s: %w", KeystoneLockfile, err)
	}
	if lf.Policies == nil {
		lf.Policies = map[string]PolicyLock{}
	}
	return &lf, nil
}

// writeLockfile serializes the lockfile back to disk under installDir.
// Always emits Version = LockfileVersion.
func writeLockfile(installDir string, lf *Lockfile) error {
	lf.Version = LockfileVersion

	out, err := yaml.Marshal(lf)
	if err != nil {
		return fmt.Errorf("marshal lockfile: %w", err)
	}

	path := filepath.Join(installDir, KeystoneLockfile)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, out, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", KeystoneLockfile, err)
	}
	return nil
}

// ensureLockfile returns the lockfile, backfilling its keystone section
// from INSTALL_PROFILE.md if the lockfile is empty (i.e., the install was
// created before the lockfile existed). The returned lockfile is not yet
// persisted — callers that mutate it must call writeLockfile.
func ensureLockfile(installDir string) (*Lockfile, error) {
	lf, err := readLockfile(installDir)
	if err != nil {
		return nil, err
	}
	if lf.Keystone.Version != "" {
		return lf, nil
	}
	if v, perr := readKeystoneVersionFromProfile(installDir); perr == nil && v != "" {
		lf.Keystone.Version = v
	}
	if agents, perr := readInstalledAgentsFromProfile(installDir); perr == nil && len(agents) > 0 {
		lf.Keystone.Agents = agents
	}
	return lf, nil
}

// hashFile returns "sha256:<hex>" for the file at path.
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}

// hashFilesUnder walks dir and returns a map of path-relative-to-installDir → hash
// for every regular file found. Used after a policy install to seed the
// lockfile, and on update to detect dirty files.
func hashFilesUnder(installDir, dir string) (map[string]string, error) {
	result := map[string]string{}
	walkRoot := filepath.Join(installDir, dir)
	err := filepath.Walk(walkRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(installDir, path)
		if relErr != nil {
			return relErr
		}
		h, herr := hashFile(path)
		if herr != nil {
			return herr
		}
		result[filepath.ToSlash(rel)] = h
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// sortedKeys returns the keys of m in sorted order — used for stable iteration
// when reporting lockfile state.
func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
