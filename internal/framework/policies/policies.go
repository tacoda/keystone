// Package policies implements the 1.0 vendored-policy flow: fetch a policy
// repo at a pinned ref into a content-addressable cache, install it under
// <charter-root>/policies/<name>/, hash every file for drift detection, and
// reset the directory on any drift.
//
// The vendor directory is gitignored at the consumer side and treated as
// read-only — chmod 0444 marks files on POSIX, but the real enforcement is
// the per-file hash check the loader runs before every cascade resolution.
package policies

// CacheDirEnv is the env var consumers can set to relocate the policy
// content-addressable cache. When empty, the cache lives under
// $XDG_CACHE_HOME / os.UserCacheDir() in `keystone/policies/`.
const CacheDirEnv = "KEYSTONE_POLICY_CACHE"

// LegacyCacheDirEnv is the pre-2.1 name of CacheDirEnv. Read as a fallback
// so users with the old export keep working until they migrate.
const LegacyCacheDirEnv = "KEYSTONE_PLUGIN_CACHE"
