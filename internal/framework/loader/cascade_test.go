package loader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tacoda/keystone/internal/framework/lockfile"
	"github.com/tacoda/keystone/internal/framework/manifest"
)

// writeFile is a tiny helper to seed an install dir for cascade verify tests.
func writeFile(t *testing.T, dir, rel string) {
	t.Helper()
	path := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

func TestVerify_CleanCascade(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "harness/policies/acme/guides/data-handling.md")
	writeFile(t, dir, "harness/guides/spec.md")

	policies := map[string]lockfile.PolicyLock{
		"acme": {
			Tier:   manifest.TierOrg,
			Strict: manifest.StrictSpec{Guides: []string{"data-handling"}},
		},
	}
	res, err := Verify(dir, policies)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if res.HasErrors() {
		t.Errorf("expected no violations, got %d: %+v", len(res.Violations), res.Violations)
	}
	if res.HasGaps() {
		t.Errorf("expected no gaps, got %d: %+v", len(res.Gaps), res.Gaps)
	}
}

func TestVerify_ProjectShadowsOrgStrict(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "harness/policies/acme/guides/data-handling.md")
	writeFile(t, dir, "harness/guides/data-handling.md") // project shadow

	policies := map[string]lockfile.PolicyLock{
		"acme": {
			Tier:   manifest.TierOrg,
			Strict: manifest.StrictSpec{Guides: []string{"data-handling"}},
		},
	}
	res, err := Verify(dir, policies)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !res.HasErrors() {
		t.Fatalf("expected violation, got none")
	}
	if got := len(res.Violations); got != 1 {
		t.Fatalf("expected 1 violation, got %d", got)
	}
	v := res.Violations[0]
	if v.Policy != "acme" || v.Kind != "guides" || v.Item != "data-handling" {
		t.Errorf("unexpected violation: %+v", v)
	}
	if len(v.ShadowPaths) != 1 || v.ShadowPaths[0] != "harness/guides/data-handling.md" {
		t.Errorf("unexpected shadow paths: %+v", v.ShadowPaths)
	}
}

func TestVerify_TeamPolicyShadowsOrgStrict(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "harness/policies/acme/guides/data-handling.md")
	writeFile(t, dir, "harness/policies/platform/guides/data-handling.md") // team shadow

	policies := map[string]lockfile.PolicyLock{
		"acme":     {Tier: manifest.TierOrg, Strict: manifest.StrictSpec{Guides: []string{"data-handling"}}},
		"platform": {Tier: manifest.TierTeam},
	}
	res, err := Verify(dir, policies)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !res.HasErrors() {
		t.Fatalf("expected violation, got none")
	}
}

func TestVerify_TeamStrict_OnlyProjectCanShadow(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "harness/policies/platform/sensors/rubocop.md")
	writeFile(t, dir, "harness/policies/other-team/sensors/rubocop.md") // team-tier peer should NOT count

	policies := map[string]lockfile.PolicyLock{
		"platform":   {Tier: manifest.TierTeam, Strict: manifest.StrictSpec{Sensors: []string{"rubocop"}}},
		"other-team": {Tier: manifest.TierTeam},
	}
	res, err := Verify(dir, policies)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if res.HasErrors() {
		t.Errorf("team-tier strict should not be shadowed by another team-tier policy; got %+v", res.Violations)
	}
}

func TestVerify_RequiredGap(t *testing.T) {
	dir := t.TempDir()
	policies := map[string]lockfile.PolicyLock{
		"acme": {
			Tier:     manifest.TierOrg,
			Required: manifest.StrictSpec{Actions: []string{"release"}},
		},
	}
	res, err := Verify(dir, policies)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !res.HasGaps() {
		t.Fatalf("expected gap, got none")
	}
	if got := len(res.Gaps); got != 1 {
		t.Fatalf("expected 1 gap, got %d", got)
	}
	g := res.Gaps[0]
	if g.Item != "release" || g.Kind != "actions" {
		t.Errorf("unexpected gap: %+v", g)
	}
}

func TestVerify_RequiredSatisfiedByProject(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "harness/actions/release.md")

	policies := map[string]lockfile.PolicyLock{
		"acme": {
			Tier:     manifest.TierOrg,
			Required: manifest.StrictSpec{Actions: []string{"release"}},
		},
	}
	res, err := Verify(dir, policies)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if res.HasGaps() {
		t.Errorf("expected no gaps when project defines the item, got %+v", res.Gaps)
	}
}
