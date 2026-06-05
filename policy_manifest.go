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

// Tier classifies a policy's authority level relative to the project.
// `org` policies sit above `team` policies in the override cascade; project
// is the leaf (the harness root itself) and is implicit — never recorded
// here. Defaults to `org` when omitted, preserving pre-tier policy behavior.
const (
	TierOrg  = "org"
	TierTeam = "team"
)

// StrictSpec lists items, by kind, for either a `strict` or `required`
// declaration on a policy. Corpus is intentionally absent — it is background
// reference loaded on-demand and not subject to the cascade.
type StrictSpec struct {
	Guides    []string `yaml:"guides,omitempty"`
	Playbooks []string `yaml:"playbooks,omitempty"`
	Actions   []string `yaml:"actions,omitempty"`
}

// IsEmpty reports whether the spec names any items.
func (s StrictSpec) IsEmpty() bool {
	return len(s.Guides) == 0 && len(s.Playbooks) == 0 && len(s.Actions) == 0
}

// Manifest describes one policy (a distributable bundle of governance
// content). Loaded from keystone-policy.yaml at the policy repo root.
//
// `strict` items are shipped by this policy and locked against override
// from lower tiers. `required` items are NOT shipped by this policy — the
// policy declares they should exist somewhere in the cascade (typically the
// project); verify surfaces missing ones so the project knows what to fill in.
type Manifest struct {
	Name        string     `yaml:"name"`
	Version     string     `yaml:"version"`
	Tier        string     `yaml:"tier,omitempty"`        // "org" (default) or "team"
	KeystoneMin string     `yaml:"keystone_min,omitempty"`
	Description string     `yaml:"description,omitempty"`
	Strict      StrictSpec `yaml:"strict,omitempty"`
	Required    StrictSpec `yaml:"required,omitempty"`
}

// ResolvedTier returns the policy's tier, applying the default if unset.
func (m *Manifest) ResolvedTier() string {
	if m.Tier == "" {
		return TierOrg
	}
	return m.Tier
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

// validate enforces required fields, name format, and tier values.
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
	switch m.Tier {
	case "", TierOrg, TierTeam:
	default:
		return fmt.Errorf("%s: tier %q must be %q or %q", PolicyManifestFile, m.Tier, TierOrg, TierTeam)
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
