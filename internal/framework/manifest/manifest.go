// Package manifest models a policy/plugin manifest and its content
// constraints. A manifest is the small declarative file at the root of every
// policy or plugin repo that names it, sets its version, and declares which
// items it ships strictly and which it requires from elsewhere in the cascade.
//
// The manifest file is keystone-plugin.json. JSON format.
package manifest

// PolicyManifestFile is the on-disk name of the manifest at the policy/plugin
// repo root.
const PolicyManifestFile = "keystone-plugin.json"

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
// reference loaded on-demand and not subject to the cascade. Sensors cascade
// across two tiers only (team → project); org policies cannot declare them.
type StrictSpec struct {
	Guides    []string `json:"guides,omitempty"`
	Playbooks []string `json:"playbooks,omitempty"`
	Actions   []string `json:"actions,omitempty"`
	Sensors   []string `json:"sensors,omitempty"`
}

// IsEmpty reports whether the spec names any items.
func (s StrictSpec) IsEmpty() bool {
	return len(s.Guides) == 0 && len(s.Playbooks) == 0 && len(s.Actions) == 0 && len(s.Sensors) == 0
}

// Manifest describes one policy (a distributable bundle of governance
// content). Loaded from keystone-plugin.json at the policy repo root.
//
// `strict` items are shipped by this policy and locked against override
// from lower tiers. `required` items are NOT shipped by this policy — the
// policy declares they should exist somewhere in the cascade (typically the
// project); verify surfaces missing ones so the project knows what to fill in.
type Manifest struct {
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	Tier        string     `json:"tier,omitempty"` // "org" (default) or "team"
	KeystoneMin string     `json:"keystone_min,omitempty"`
	Description string     `json:"description,omitempty"`
	Strict      StrictSpec `json:"strict,omitempty"`
	Required    StrictSpec `json:"required,omitempty"`
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
