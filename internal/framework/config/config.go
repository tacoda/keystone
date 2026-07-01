// Package config holds project-wide settings that are stable across the
// lifetime of an install: the charter directory name, framework version
// pin, and (in later phases) the policy tree declared in keystone.json.
//
// At Phase 2 only the charter root is configurable here. Phase 3 will fold
// the full keystone.json schema into this package.
package config

// DefaultCharterRoot is the single standardized directory that holds the
// charter inside a consumer's repo and feeds every agent. At 3.0 this is
// `.charter` — one agent-neutral location (the name describes the thing,
// not the tool). The primitives, the generated INDEX.json, the lockfile,
// and the document workspace all live under it; projection reads it and
// writes each host surface (.claude/, .cursor/, AGENTS.md, …).
//
// This is the one place to change the charter location. `keystone.json`
// stays at the repo root as the entry-point config.
const DefaultCharterRoot = ".charter"

// LockfileName is the basename of the per-install state record, written
// at <charter-root>/lockfile.json. At 3.0 this resolves to
// `.charter/lockfile.json`.
const LockfileName = "lockfile.json"

// IndexName is the basename of the generated primitive descriptor index,
// written at <charter-root>/INDEX.json.
const IndexName = "INDEX.json"

// IndexLiteName is the basename of the cheap-discovery sibling index
// (kind/id/description only). Written alongside INDEX.json by every
// `keystone index` run. Agents reference this for first-pass browsing;
// INDEX.json is opened only when a path, glob, or trigger is needed to
// activate a specific primitive.
const IndexLiteName = "INDEX.lite.json"

// KeystoneDir returns the directory that holds the lockfile, INDEX.json,
// and document workspace. At 3.0 the charter root IS that umbrella — they
// live directly under `.charter/` — so this returns the charter root
// itself.
//
// Pre-3.0 installs nested the charter under `.charter/` (the umbrella
// was the parent of the charter root). Those paths are pinned inside the
// v2.x migrations and the `keystone migrate` legacy-layout detection, not
// derived here, so this function need not special-case them.
func KeystoneDir(charterRoot string) string {
	if charterRoot == "" {
		charterRoot = DefaultCharterRoot
	}
	return charterRoot
}
