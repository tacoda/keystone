package plugins

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Cached is the result of a successful Fetch. Dir holds the cache entry
// the plugin was unpacked into; ResolvedSHA is the exact commit checked
// out at fetch time.
type Cached struct {
	Dir         string
	ResolvedSHA string
}

// Fetch resolves gitURL at version into the content-addressable cache and
// returns the cached directory. gitURL is expected to be a full URL
// (https://, file://, git@host:path); callers using shorthand source
// strings should run them through config.ExpandSource first.
//
// version is passed to `git clone --branch`, which accepts tags and branch
// names. For an arbitrary commit SHA, the fetch falls back to a full clone
// + checkout.
//
// Cache key is sha256(gitURL + "@" + version), so two different shorthand
// forms that expand to the same URL share one cache entry.
func Fetch(gitURL, version string) (*Cached, error) {
	if gitURL == "" {
		return nil, fmt.Errorf("plugins.Fetch: empty gitURL")
	}
	if version == "" {
		return nil, fmt.Errorf("plugins.Fetch: empty version")
	}

	cacheDir, err := cachePath(gitURL, version)
	if err != nil {
		return nil, err
	}

	if sha, ok := readCacheSHA(cacheDir); ok {
		return &Cached{Dir: cacheDir, ResolvedSHA: sha}, nil
	}

	if err := os.MkdirAll(filepath.Dir(cacheDir), 0o755); err != nil {
		return nil, fmt.Errorf("create cache dir: %w", err)
	}

	tmp, err := os.MkdirTemp(filepath.Dir(cacheDir), "fetching-")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(tmp) }

	// Shallow clone by ref first; fall back to full clone for commit SHAs.
	if err := runGit("", "clone", "--quiet", "--branch", version, "--depth", "1", gitURL, tmp); err != nil {
		cleanup()
		tmp, err = os.MkdirTemp(filepath.Dir(cacheDir), "fetching-")
		if err != nil {
			return nil, fmt.Errorf("create temp dir: %w", err)
		}
		cleanup = func() { _ = os.RemoveAll(tmp) }
		if err := runGit("", "clone", "--quiet", gitURL, tmp); err != nil {
			cleanup()
			return nil, fmt.Errorf("git clone %s: %w", gitURL, err)
		}
		if err := runGit(tmp, "checkout", "--quiet", version); err != nil {
			cleanup()
			return nil, fmt.Errorf("git checkout %s: %w", version, err)
		}
	}

	sha, err := gitRevParseHead(tmp)
	if err != nil {
		cleanup()
		return nil, err
	}
	if err := writeCacheSHA(tmp, sha); err != nil {
		cleanup()
		return nil, err
	}

	// Atomic move into the cache. If the rename races with another fetch
	// for the same key, the loser cleans up and uses the winner's entry.
	if err := os.Rename(tmp, cacheDir); err != nil {
		cleanup()
		if sha, ok := readCacheSHA(cacheDir); ok {
			return &Cached{Dir: cacheDir, ResolvedSHA: sha}, nil
		}
		return nil, fmt.Errorf("move to cache: %w", err)
	}
	return &Cached{Dir: cacheDir, ResolvedSHA: sha}, nil
}

// cachePath returns the absolute directory under the keystone plugin
// cache for gitURL@version.
func cachePath(gitURL, version string) (string, error) {
	root, err := cacheRoot()
	if err != nil {
		return "", err
	}
	key := sha256.Sum256([]byte(gitURL + "@" + version))
	return filepath.Join(root, hex.EncodeToString(key[:])), nil
}

// cacheRoot returns the directory holding all cache entries for plugins.
// KEYSTONE_PLUGIN_CACHE overrides the default, mainly for tests.
func cacheRoot() (string, error) {
	if v := os.Getenv(CacheDirEnv); v != "" {
		return v, nil
	}
	base, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("locate user cache dir: %w", err)
	}
	return filepath.Join(base, "keystone", "plugins"), nil
}

// readCacheSHA returns the SHA recorded inside an existing cache entry.
// Used to skip refetching when the cache already has a hit.
func readCacheSHA(cacheDir string) (string, bool) {
	data, err := os.ReadFile(filepath.Join(cacheDir, ".keystone-resolved-sha"))
	if err != nil {
		return "", false
	}
	sha := strings.TrimSpace(string(data))
	if sha == "" {
		return "", false
	}
	return sha, true
}

// writeCacheSHA stores the resolved SHA inside a cache entry for later
// reads to pick up.
func writeCacheSHA(cacheDir, sha string) error {
	return os.WriteFile(filepath.Join(cacheDir, ".keystone-resolved-sha"), []byte(sha+"\n"), 0o644)
}

// runGit runs `git <args...>` with cwd set to dir (empty = inherit cwd).
// Stderr is captured and included in the error on failure.
func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// gitRevParseHead returns the resolved commit SHA at HEAD of the repo at dir.
func gitRevParseHead(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse HEAD: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
