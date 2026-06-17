// Package plugins implements the 1.0 vendored-plugin flow: fetch a plugin
// repo at a pinned ref into a content-addressable cache, install it under
// <harness-root>/plugins/<name>/, hash every file for drift detection, and
// reset the directory on any drift.
//
// The vendor directory is gitignored at the consumer side and treated as
// read-only — chmod 0444 marks files on POSIX, but the real enforcement is
// the per-file hash check the loader runs before every cascade resolution.
package policies

// CacheDirEnv is the env var consumers can set to relocate the plugin
// content-addressable cache. When empty, the cache lives under
// $XDG_CACHE_HOME / os.UserCacheDir() in `keystone/plugins/`.
const CacheDirEnv = "KEYSTONE_PLUGIN_CACHE"
