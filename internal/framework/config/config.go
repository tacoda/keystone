// Package config holds project-wide settings that are stable across the
// lifetime of an install: the harness directory name, framework version
// pin, and (in later phases) the plugin tree declared in keystone.json.
//
// At Phase 2 only the harness root is configurable here. Phase 3 will fold
// the full keystone.json schema into this package.
package config

// DefaultHarnessRoot is the directory name where keystone installs the
// harness inside a consumer's repo. Configurable per-install via the
// --harness-root flag at `keystone init`; downstream commands accept the
// same flag (or, post-Phase 3, pick it up from keystone.json).
//
// Teams that want a different name (e.g. "agent-rules" or "playbook")
// pass --harness-root <name> at init time and re-pass it on every
// subsequent command. Default falls back here.
const DefaultHarnessRoot = "harness"

// LockfileName is the basename of the per-install state record, written
// at <harness-root>/keystone.lock.json.
const LockfileName = "keystone.lock.json"
