// Package config holds project-wide settings that are stable across the
// lifetime of an install: the harness directory name, framework version
// pin, and (in later phases) the plugin tree declared in keystone.json.
//
// At Phase 2 only the harness root is configurable here. Phase 3 will fold
// the full keystone.json schema into this package.
package config

// DefaultHarnessRoot is the relative path where keystone installs the
// harness inside a consumer's repo. At 2.0 this is `.keystone/harness` —
// the harness lives under a `.keystone/` parent dir that also holds the
// lockfile and generated INDEX.json, so both the CLI and keystone-mcp
// can discover the harness with one well-known prefix.
//
// Configurable per-install via the --harness-root flag at
// `keystone init`; downstream commands accept the same flag or pick it
// up from keystone.json. The parent directory of harnessRoot is the
// "keystone dir" — see KeystoneDir.
const DefaultHarnessRoot = ".keystone/harness"

// LockfileName is the basename of the per-install state record, written
// at <keystone-dir>/lockfile.json (i.e. one level above the harness
// root). At 2.0 this resolves to `.keystone/lockfile.json`.
const LockfileName = "lockfile.json"

// IndexName is the basename of the generated primitive descriptor index,
// written at <keystone-dir>/INDEX.json — alongside the lockfile, not
// inside the harness tree.
const IndexName = "INDEX.json"

// KeystoneDir returns the parent directory of the configured harness
// root — the `.keystone/` umbrella that holds the lockfile, the
// INDEX.json, and the harness/ subtree. For the default
// (`.keystone/harness`) it returns `.keystone`. When the user has
// configured a flat harness root (no separator), it returns the
// project-root sentinel `.`.
func KeystoneDir(harnessRoot string) string {
	if harnessRoot == "" {
		harnessRoot = DefaultHarnessRoot
	}
	parent := pathDir(harnessRoot)
	if parent == "" {
		return "."
	}
	return parent
}

// pathDir is filepath.Dir-equivalent for forward-slash POSIX paths only
// (the harness root is always declared in POSIX form in keystone.json).
// Imported inline to avoid pulling filepath into the config package's
// no-os surface.
func pathDir(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' {
			return p[:i]
		}
	}
	return ""
}
