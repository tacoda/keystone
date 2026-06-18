// Package manifest models a policy manifest. A manifest is the small
// declarative file at the root of every policy repo that names it, sets its
// version, and declares which items it ships strictly and which it requires
// from elsewhere in the cascade.
//
// The manifest file is keystone-policy.json. JSON format.
package manifest

// PolicyManifestFile is the on-disk name of the manifest at the policy repo
// root.
const PolicyManifestFile = "keystone-policy.json"

// StrictSpec lists items, by kind, for either a `strict` or `required`
// declaration on a policy. Corpus is intentionally absent — it is background
// reference loaded on-demand and not subject to the cascade.
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

// Manifest describes one policy. Loaded from keystone-policy.json at the
// policy repo root.
//
// `strict` items are shipped by this policy and locked absolutely — nothing
// else in the cascade (project or any other policy) can override them.
// `required` items are NOT shipped by this policy; the policy declares they
// should be provided by some layer outer than itself (or by the project).
// `keystone verify` surfaces missing required items as advisory gaps.
type Manifest struct {
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	KeystoneMin string     `json:"keystone_min,omitempty"`
	Description string     `json:"description,omitempty"`
	Strict      StrictSpec `json:"strict,omitempty"`
	Required    StrictSpec `json:"required,omitempty"`
}

// Namespace returns the on-disk directory name used inside harness/policies/
// for this policy's content. Always equal to the manifest name.
func (m *Manifest) Namespace() string {
	return m.Name
}
