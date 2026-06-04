package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// PolicyManifestFile is the on-disk name of the policy manifest at the policy
// repo root.
const PolicyManifestFile = "keystone-policy.yaml"

// PolicyContentRoot is the directory inside a policy repo that holds the
// files to be copied into a consumer install. Everything outside this prefix
// (README.md at repo root, the manifest itself, .git, etc.) is ignored.
const PolicyContentRoot = "policy"

// Manifest describes one org policy (a.k.a. "pack" — a distributable bundle
// of governance content). Loaded from keystone-policy.yaml at the policy
// repo root.
type Manifest struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	KeystoneMin string `yaml:"keystone_min,omitempty"`
	Description string `yaml:"description,omitempty"`
}

// Namespace returns the on-disk directory name used inside harness/policies/
// for this policy's content. Always equal to the manifest name.
func (m *Manifest) Namespace() string {
	return m.Name
}

// loadManifest reads and parses keystone-policy.yaml from policyRoot.
func loadManifest(policyRoot string) (*Manifest, error) {
	path := filepath.Join(policyRoot, PolicyManifestFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", PolicyManifestFile, err)
	}
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse %s: %w", PolicyManifestFile, err)
	}
	if err := m.validate(); err != nil {
		return nil, err
	}
	return &m, nil
}

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)

// validate enforces required fields and name format.
func (m *Manifest) validate() error {
	if m.Name == "" {
		return fmt.Errorf("%s: missing required field 'name'", PolicyManifestFile)
	}
	if !namePattern.MatchString(m.Name) {
		return fmt.Errorf("%s: name %q must match %s", PolicyManifestFile, m.Name, namePattern)
	}
	if m.Version == "" {
		return fmt.Errorf("%s: missing required field 'version'", PolicyManifestFile)
	}
	return nil
}

// validatePolicyContent walks policyRoot/policy/ and ensures every regular
// file lives under policy/harness/policies/<namespace>/. Returns the list of
// relative file paths (within policyRoot) that will be copied, or an error
// if any file is outside the allowed namespace.
func validatePolicyContent(policyRoot string, m *Manifest) ([]string, error) {
	contentRoot := filepath.Join(policyRoot, PolicyContentRoot)
	info, err := os.Stat(contentRoot)
	if err != nil {
		return nil, fmt.Errorf("policy must contain a %s/ directory: %w", PolicyContentRoot, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", PolicyContentRoot)
	}

	allowedPrefix := filepath.Join(PolicyContentRoot, "harness", "policies", m.Namespace())
	var files []string
	var stray []string

	err = filepath.WalkDir(contentRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(policyRoot, path)
		if relErr != nil {
			return relErr
		}
		if !strings.HasPrefix(rel, allowedPrefix+string(filepath.Separator)) {
			stray = append(stray, rel)
			return nil
		}
		files = append(files, rel)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(stray) > 0 {
		return nil, fmt.Errorf(
			"policy %q writes files outside its allowed namespace %s/:\n  %s",
			m.Name, allowedPrefix, strings.Join(stray, "\n  "),
		)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("policy %q contains no files under %s/", m.Name, allowedPrefix)
	}
	return files, nil
}
