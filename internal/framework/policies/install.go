package policies

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// PolicyRoot is the directory under <project>/<harnessRoot>/ where every
// installed plugin lives. Gitignored at the consumer side; managed by the
// vendor flow, not by users.
const PolicyRoot = "policies"

// readOnlyMode is the file permission applied after install on POSIX.
// Best-effort UX hint that vendored files are not meant to be edited; the
// real enforcement is the hash check in Verify.
const readOnlyMode = 0o444

// Installed describes the result of a successful Install: per-file hashes
// the lockfile records, plus the manifest's declared name/version (for
// audit) when the plugin shipped one.
type Installed struct {
	Files         map[string]string
	PolicyVersion string // value from keystone-policy.json's `version`, if present
	PolicyName    string // value from keystone-policy.json's `name`, if present
}

// pluginManifestFile is the basename of a plugin's manifest at the root of
// its content tree. Local to this package so plugins doesn't depend on the
// manifest package.
const pluginManifestFile = "keystone-policy.json"

// pluginManifest is the minimal subset of keystone-policy.json the
// installer reads. The full schema lives in docs/schemas/.
type pluginManifest struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Install copies the cached plugin tree into <projectDir>/<harnessRoot>/plugins/<name>/,
// computes per-file SHA-256 hashes for the lockfile, and chmods files
// read-only on POSIX (best-effort).
//
// `cached.Dir` is the result of Fetch; we trust the cache and copy
// everything except the .keystone-resolved-sha sentinel into the
// vendor directory.
//
// If the destination already exists, it is removed first — Install is the
// "fresh write" path used by both first-install and drift-reset.
func Install(cached *Cached, name, projectDir, harnessRoot string) (*Installed, error) {
	if cached == nil {
		return nil, fmt.Errorf("policies.Install: nil Cached")
	}
	if name == "" {
		return nil, fmt.Errorf("policies.Install: empty name")
	}
	if harnessRoot == "" {
		return nil, fmt.Errorf("policies.Install: empty harnessRoot")
	}

	target := pluginDir(projectDir, harnessRoot, name)
	if err := os.RemoveAll(target); err != nil {
		return nil, fmt.Errorf("clear target: %w", err)
	}
	if err := os.MkdirAll(target, 0o755); err != nil {
		return nil, fmt.Errorf("create target: %w", err)
	}

	files := map[string]string{}
	var manifest *pluginManifest

	err := filepath.WalkDir(cached.Dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(cached.Dir, path)
		if relErr != nil {
			return relErr
		}
		if rel == "." {
			return nil
		}
		// Skip the cache sentinel and .git internals — neither belongs in the
		// vendor directory.
		if rel == ".keystone-resolved-sha" {
			return nil
		}
		if rel == ".git" || strings.HasPrefix(rel, ".git"+string(filepath.Separator)) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		// 2.0 constraint: policies extend FRAMEWORK abstractions only.
		// Agent-side abstractions (rule, skill, subagent, command) are
		// project-owned — a policy that ships them is rejected.
		if top := topSegment(rel); isAgentAbstractionDir(top) {
			return fmt.Errorf("policy %q ships an agent-abstraction directory %q; policies may only extend framework abstractions (guides, corpus, sensors, actions, playbooks, adapters)", name, top)
		}

		destPath := filepath.Join(target, rel)
		if d.IsDir() {
			return os.MkdirAll(destPath, 0o755)
		}

		hash, err := copyFile(path, destPath)
		if err != nil {
			return err
		}

		// Record the hash keyed by path relative to the project root, so the
		// lockfile entry lines up with how Verify walks the tree.
		relFromProject := filepath.ToSlash(filepath.Join(harnessRoot, PolicyRoot, name, rel))
		files[relFromProject] = hash

		// Capture the manifest's declared name/version on the way past.
		if rel == pluginManifestFile {
			data, err := os.ReadFile(destPath)
			if err == nil {
				var m pluginManifest
				if err := json.Unmarshal(data, &m); err == nil {
					manifest = &m
				}
			}
		}

		if runtime.GOOS != "windows" {
			_ = os.Chmod(destPath, readOnlyMode)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("copy plugin tree: %w", err)
	}

	installed := &Installed{Files: files}
	if manifest != nil {
		installed.PolicyName = manifest.Name
		installed.PolicyVersion = manifest.Version
	}
	return installed, nil
}

// pluginDir returns the absolute path of the install directory for a
// single plugin.
func pluginDir(projectDir, harnessRoot, name string) string {
	return filepath.Join(projectDir, harnessRoot, PolicyRoot, name)
}

// copyFile streams src to dst and returns the "sha256:<hex>" hash of the
// content. The destination directory is created if missing.
func copyFile(src, dst string) (string, error) {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return "", err
	}
	in, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	if _, err := io.Copy(io.MultiWriter(out, h), in); err != nil {
		out.Close()
		return "", err
	}
	if err := out.Close(); err != nil {
		return "", err
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}

// topSegment returns the first path component of a slash-or-OS-separator
// relative path. For "guides/foo/bar.md" it returns "guides"; for a
// top-level file like "README.md" it returns "README.md".
func topSegment(rel string) string {
	rel = filepath.ToSlash(rel)
	if i := strings.IndexByte(rel, '/'); i >= 0 {
		return rel[:i]
	}
	return rel
}

// isAgentAbstractionDir reports whether seg names a directory that
// holds agent-side primitives (rule, skill, subagent, command).
// Policies are forbidden from shipping these — agent abstractions are
// project-owned.
func isAgentAbstractionDir(seg string) bool {
	switch seg {
	case "rules", "skills", "agents", "commands":
		return true
	}
	return false
}
