package loader

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tacoda/keystone/internal/framework/lockfile"
	"github.com/tacoda/keystone/internal/framework/manifest"
)

// VerifyResult is the outcome of a cascade verification against an install
// directory. Violations are hard errors (strict cascade broken). Gaps are
// informational — items a policy says should exist somewhere but no tier
// has provided.
type VerifyResult struct {
	Violations []ShadowViolation
	Gaps       []MissingRequired
}

// HasErrors reports whether any strict rule was violated.
func (r VerifyResult) HasErrors() bool { return len(r.Violations) > 0 }

// HasGaps reports whether any required item is unmet.
func (r VerifyResult) HasGaps() bool { return len(r.Gaps) > 0 }

// ShadowViolation reports one strict item being overridden by a lower tier.
// Paths are recorded relative to the install dir, slash-separated.
type ShadowViolation struct {
	Policy      string   // name of the policy that declared the item strict
	PolicyTier  string   // "org" or "team"
	Kind        string   // "guides" / "playbooks" / "actions" / "sensors"
	Item        string   // basename (no .md)
	ShadowPaths []string // file paths that violate the strict rule
}

// String renders the violation for terminal output.
func (v ShadowViolation) String() string {
	return fmt.Sprintf("policy %q (tier %s) marks %s/%s strict — overridden by:\n    %s",
		v.Policy, v.PolicyTier, v.Kind, v.Item, strings.Join(v.ShadowPaths, "\n    "))
}

// MissingRequired reports a `required` item that no tier in the cascade has
// defined. The project is on the hook to provide it.
type MissingRequired struct {
	Policy     string // name of the policy that declared the item required
	PolicyTier string // "org" or "team"
	Kind       string // "guides" / "playbooks" / "actions" / "sensors"
	Item       string // basename (no .md)
}

// String renders the gap for terminal output.
func (m MissingRequired) String() string {
	return fmt.Sprintf("policy %q (tier %s) requires %s/%s — define it at harness/%s/%s.md",
		m.Policy, m.PolicyTier, m.Kind, m.Item, m.Kind, m.Item)
}

// Verify walks every installed policy's strict + required items and reports
// strict-cascade shadow violations (hard) and required-item gaps (advisory).
// The returned *VerifyResult is always non-nil. Errors here are I/O failures,
// not policy violations.
func Verify(installDir string, policies map[string]lockfile.PolicyLock) (*VerifyResult, error) {
	res := &VerifyResult{}

	for _, policyName := range sortedPolicyNames(policies) {
		lock := policies[policyName]
		tier := lock.ResolvedTier()
		for _, kind := range []string{"guides", "playbooks", "actions", "sensors"} {
			for _, item := range strictItemsFor(lock.Strict, kind) {
				paths, walkErr := findShadowing(installDir, kind, item, tier, policies, policyName)
				if walkErr != nil {
					return nil, walkErr
				}
				if len(paths) == 0 {
					continue
				}
				res.Violations = append(res.Violations, ShadowViolation{
					Policy:      policyName,
					PolicyTier:  tier,
					Kind:        kind,
					Item:        item,
					ShadowPaths: paths,
				})
			}
			for _, item := range strictItemsFor(lock.Required, kind) {
				defined, walkErr := isItemDefined(installDir, kind, item, policies)
				if walkErr != nil {
					return nil, walkErr
				}
				if defined {
					continue
				}
				res.Gaps = append(res.Gaps, MissingRequired{
					Policy:     policyName,
					PolicyTier: tier,
					Kind:       kind,
					Item:       item,
				})
			}
		}
	}
	return res, nil
}

// isItemDefined returns true when any tier — project tree or any installed
// policy — has a file named `<item>.md` under the kind subtree.
func isItemDefined(installDir, kind, item string, policies map[string]lockfile.PolicyLock) (bool, error) {
	projectPaths, err := findItemInTree(installDir, filepath.Join("harness", kind), item)
	if err != nil {
		return false, err
	}
	if len(projectPaths) > 0 {
		return true, nil
	}
	for _, name := range sortedPolicyNames(policies) {
		policyPaths, err := findItemInTree(installDir, filepath.Join("harness", "policies", name, kind), item)
		if err != nil {
			return false, err
		}
		if len(policyPaths) > 0 {
			return true, nil
		}
	}
	return false, nil
}

// strictItemsFor returns the basenames listed for the given kind.
func strictItemsFor(s manifest.StrictSpec, kind string) []string {
	switch kind {
	case "guides":
		return s.Guides
	case "playbooks":
		return s.Playbooks
	case "actions":
		return s.Actions
	case "sensors":
		return s.Sensors
	}
	return nil
}

// findShadowing returns every project- or lower-tier-policy file whose
// basename matches `<item>.md` under the kind subtree. The declaring policy's
// own files are excluded (a policy never shadows itself).
//
// For org-tier strict rules, team-tier policies + project paths are
// candidates. For team-tier strict rules, only project paths are.
func findShadowing(installDir, kind, item, declaringTier string, policies map[string]lockfile.PolicyLock, declaringName string) ([]string, error) {
	var hits []string

	projectPaths, err := findItemInTree(installDir, filepath.Join("harness", kind), item)
	if err != nil {
		return nil, err
	}
	hits = append(hits, projectPaths...)

	if declaringTier == manifest.TierOrg {
		for _, name := range sortedPolicyNames(policies) {
			if name == declaringName {
				continue
			}
			peer := policies[name]
			if peer.ResolvedTier() != manifest.TierTeam {
				continue
			}
			teamPaths, err := findItemInTree(installDir, filepath.Join("harness", "policies", name, kind), item)
			if err != nil {
				return nil, err
			}
			hits = append(hits, teamPaths...)
		}
	}
	return hits, nil
}

// findItemInTree walks subdir under installDir and returns relative paths of
// every file named `<item>.md`. Returns nil (not an error) when subdir is
// absent.
func findItemInTree(installDir, subdir, item string) ([]string, error) {
	root := filepath.Join(installDir, subdir)
	want := item + ".md"

	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil, nil
	}

	var hits []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Base(path) != want {
			return nil
		}
		rel, relErr := filepath.Rel(installDir, path)
		if relErr != nil {
			return relErr
		}
		hits = append(hits, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return hits, nil
}

// sortedPolicyNames returns deterministic iteration order for a policy map.
func sortedPolicyNames(m map[string]lockfile.PolicyLock) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
