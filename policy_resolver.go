package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// PolicyRef is a parsed `--policy` argument. In v1 only the git transport is
// supported: `git+<url>` with an optional `#<rev>` suffix.
type PolicyRef struct {
	Raw    string // original user-supplied string
	Scheme string // "git" in v1
	URL    string // transport-specific URL (e.g., "https://github.com/acme/policy.git")
	Rev    string // tag, branch, or SHA; empty means "default branch HEAD"
}

// parsePolicyRef parses a user-supplied --policy argument. Format:
//
//	git+<url>[#<rev>]
//
// Examples:
//
//	git+https://github.com/acme/policy.git
//	git+https://github.com/acme/policy.git#v1.2.0
//	git+ssh://git@github.com/acme/policy.git#main
func parsePolicyRef(raw string) (*PolicyRef, error) {
	if raw == "" {
		return nil, fmt.Errorf("empty policy ref")
	}
	if !strings.HasPrefix(raw, "git+") {
		return nil, fmt.Errorf("unsupported policy ref %q: only `git+<url>[#<rev>]` is supported in v1", raw)
	}
	body := strings.TrimPrefix(raw, "git+")

	ref := &PolicyRef{Raw: raw, Scheme: "git"}
	if i := strings.LastIndex(body, "#"); i >= 0 {
		ref.URL = body[:i]
		ref.Rev = body[i+1:]
	} else {
		ref.URL = body
	}
	if ref.URL == "" {
		return nil, fmt.Errorf("policy ref %q: empty URL", raw)
	}
	return ref, nil
}

// ResolvedPolicy is the result of fetching a policy to local disk. LocalDir
// is a temp directory owned by the caller — clean it up with os.RemoveAll
// when done.
type ResolvedPolicy struct {
	Ref         *PolicyRef
	LocalDir    string // temp dir containing the policy repo root
	ResolvedSHA string // exact commit checked out
}

// resolvePolicy dispatches by scheme. Only git is implemented in v1.
func resolvePolicy(ref *PolicyRef) (*ResolvedPolicy, error) {
	switch ref.Scheme {
	case "git":
		return resolveGit(ref)
	default:
		return nil, fmt.Errorf("scheme %q not supported", ref.Scheme)
	}
}

// resolveGit clones the policy repo to a temp dir and checks out the requested
// revision. Shallow clone is attempted first (works for tags and branches);
// falls back to full clone + checkout for arbitrary commit SHAs.
func resolveGit(ref *PolicyRef) (*ResolvedPolicy, error) {
	dir, err := os.MkdirTemp("", "keystone-policy-")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}

	cleanup := func() { _ = os.RemoveAll(dir) }

	if ref.Rev != "" {
		if err := runGit("", "clone", "--quiet", "--branch", ref.Rev, "--depth", "1", ref.URL, dir); err == nil {
			sha, sherr := gitRevParseHead(dir)
			if sherr != nil {
				cleanup()
				return nil, sherr
			}
			return &ResolvedPolicy{Ref: ref, LocalDir: dir, ResolvedSHA: sha}, nil
		}
		// shallow clone failed — could be a SHA. Fall through to full clone.
		cleanup()
		dir, err = os.MkdirTemp("", "keystone-policy-")
		if err != nil {
			return nil, fmt.Errorf("create temp dir: %w", err)
		}
		cleanup = func() { _ = os.RemoveAll(dir) }
	}

	if err := runGit("", "clone", "--quiet", ref.URL, dir); err != nil {
		cleanup()
		return nil, fmt.Errorf("git clone %s: %w", ref.URL, err)
	}
	if ref.Rev != "" {
		if err := runGit(dir, "checkout", "--quiet", ref.Rev); err != nil {
			cleanup()
			return nil, fmt.Errorf("git checkout %s: %w", ref.Rev, err)
		}
	}

	sha, err := gitRevParseHead(dir)
	if err != nil {
		cleanup()
		return nil, err
	}
	return &ResolvedPolicy{Ref: ref, LocalDir: dir, ResolvedSHA: sha}, nil
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
