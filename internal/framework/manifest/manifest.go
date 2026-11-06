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
// declaration on a policy. References are absent: they load on demand and
// are not subject to the cascade. The 4.x keys (guides, playbooks,
// actions, sensors) stay readable and lock the same items in 5.0 folders.
type StrictSpec struct {
	Rules    []string `json:"rules,omitempty"`
	Skills   []string `json:"skills,omitempty"`
	Commands []string `json:"commands,omitempty"`
	Hooks    []string `json:"hooks,omitempty"`

	Guides    []string `json:"guides,omitempty"`    // 4.x: rules
	Playbooks []string `json:"playbooks,omitempty"` // 4.x: skills
	Actions   []string `json:"actions,omitempty"`   // 4.x: commands
	Sensors   []string `json:"sensors,omitempty"`   // 4.x: hooks
}

// ByKey is the spec as key → items, for every key that names items.
func (s StrictSpec) ByKey() map[string][]string {
	out := map[string][]string{}
	for key, items := range map[string][]string{
		"rules": s.Rules, "skills": s.Skills, "commands": s.Commands, "hooks": s.Hooks,
		"guides": s.Guides, "playbooks": s.Playbooks, "actions": s.Actions, "sensors": s.Sensors,
	} {
		if len(items) > 0 {
			out[key] = items
		}
	}
	return out
}

// IsEmpty reports whether the spec names any items.
func (s StrictSpec) IsEmpty() bool { return len(s.ByKey()) == 0 }

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

// Namespace returns the on-disk directory name used inside charter/policies/
// for this policy's content. Always equal to the manifest name.
func (m *Manifest) Namespace() string {
	return m.Name
}
