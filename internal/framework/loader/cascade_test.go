package loader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tacoda/keystone/internal/framework/config"
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
	writeFile(t, dir, ".agents/rules/spec.md") // project file, unrelated

	cfg := &config.ProjectConfig{
		Version: config.SchemaVersion,
		Policies: []config.PolicyNode{
			{
				Name:    "acme",
				Source:  "acme/policies",
				Version: "v1",
				Strict:  map[string][]string{"guides": {"data-handling"}},
			},
		},
	}
	res, err := Verify(dir, cfg, nil)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if res.HasErrors() {
		t.Errorf("expected no violations, got %d: %+v", len(res.Violations), res.Violations)
	}
}

func TestVerify_ProjectShadowsStrict(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".agents/rules/data-handling.md") // project shadow

	cfg := &config.ProjectConfig{
		Version: config.SchemaVersion,
		Policies: []config.PolicyNode{
			{
				Name:    "acme",
				Source:  "acme/policies",
				Version: "v1",
				Strict:  map[string][]string{"guides": {"data-handling"}},
			},
		},
	}
	res, err := Verify(dir, cfg, nil)
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
	if v.Policy != "acme" || v.Port != "guides" || v.Item != "data-handling" {
		t.Errorf("unexpected violation: %+v", v)
	}
	if v.PathContext != "acme" {
		t.Errorf("PathContext = %q, want %q", v.PathContext, "acme")
	}
	if len(v.ShadowPaths) != 1 || v.ShadowPaths[0] != ".agents/rules/data-handling.md" {
		t.Errorf("unexpected shadow paths: %+v", v.ShadowPaths)
	}
}

func TestVerify_NestedPolicyPathContext(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".agents/hooks/rubocop.md")

	cfg := &config.ProjectConfig{
		Version: config.SchemaVersion,
		Policies: []config.PolicyNode{
			{
				Name:    "acme-org",
				Source:  "acme/org",
				Version: "v1",
				Children: []config.PolicyNode{
					{
						Name:    "acme-platform",
						Source:  "acme/platform",
						Version: "v1",
						Strict:  map[string][]string{"sensors": {"rubocop"}},
					},
				},
			},
		},
	}
	res, err := Verify(dir, cfg, nil)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !res.HasErrors() {
		t.Fatalf("expected violation, got none")
	}
	v := res.Violations[0]
	if v.PathContext != "acme-org > acme-platform" {
		t.Errorf("PathContext = %q, want %q", v.PathContext, "acme-org > acme-platform")
	}
}

func TestVerify_DepthGate_NestedPolicyCannotShipSensors(t *testing.T) {
	dir := t.TempDir()
	// Vendored sensor file shipped by the nested policy.
	writeFile(t, dir, ".keystone/policies/acme-platform/sensors/rubocop.md")

	cfg := &config.ProjectConfig{
		Version: config.SchemaVersion,
		Policies: []config.PolicyNode{
			{
				Name:    "acme-org",
				Source:  "acme/org",
				Version: "v1",
				Children: []config.PolicyNode{
					{
						Name:    "acme-platform",
						Source:  "acme/platform",
						Version: "v1",
						Strict:  map[string][]string{"sensors": {"rubocop"}},
					},
				},
			},
		},
	}
	expected := map[string]map[string]string{
		"acme-platform": {".keystone/policies/acme-platform/sensors/rubocop.md": "sha256:deadbeef"},
	}
	res, err := Verify(dir, cfg, expected)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !res.HasErrors() {
		t.Fatalf("expected errors, got none")
	}
	if len(res.DepthViolations) != 1 {
		t.Fatalf("DepthViolations = %d, want 1: %+v", len(res.DepthViolations), res.DepthViolations)
	}
	dv := res.DepthViolations[0]
	if dv.Policy != "acme-platform" {
		t.Errorf("Policy = %q, want %q", dv.Policy, "acme-platform")
	}
	if dv.Depth != 1 {
		t.Errorf("Depth = %d, want 1", dv.Depth)
	}
	if dv.PathContext != "acme-org > acme-platform" {
		t.Errorf("PathContext = %q, want %q", dv.PathContext, "acme-org > acme-platform")
	}
	if len(dv.StrictSensors) != 1 || dv.StrictSensors[0] != "rubocop" {
		t.Errorf("StrictSensors = %v, want [rubocop]", dv.StrictSensors)
	}
	if len(dv.VendoredSensors) != 1 || dv.VendoredSensors[0] != ".keystone/policies/acme-platform/sensors/rubocop.md" {
		t.Errorf("VendoredSensors = %v, want [charter/policies/acme-platform/sensors/rubocop.md]", dv.VendoredSensors)
	}
}

func TestVerify_DepthGate_TopLevelPolicyMayShipSensors(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".keystone/policies/acme-org/sensors/rubocop.md")

	cfg := &config.ProjectConfig{
		Version: config.SchemaVersion,
		Policies: []config.PolicyNode{
			{
				Name:    "acme-org",
				Source:  "acme/org",
				Version: "v1",
				Strict:  map[string][]string{"sensors": {"rubocop"}},
			},
		},
	}
	expected := map[string]map[string]string{
		"acme-org": {".keystone/policies/acme-org/sensors/rubocop.md": "sha256:deadbeef"},
	}
	res, err := Verify(dir, cfg, expected)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if len(res.DepthViolations) != 0 {
		t.Errorf("DepthViolations = %d, want 0 (top-level policy may ship sensors): %+v", len(res.DepthViolations), res.DepthViolations)
	}
}

func TestVerify_RequiredGap_ProjectSatisfies(t *testing.T) {
	dir := t.TempDir()
	// Project ships actions/release-notes.md → satisfies tacoda-org's required claim.
	writeFile(t, dir, ".agents/commands/release-notes.md")
	// Drop a minimal manifest for the installed policy that declares the required item.
	writeFile(t, dir, ".keystone/policies/tacoda-org/dummy.md") // ensure dir exists
	manifest := `{
  "name": "tacoda-org",
  "version": "1.0.0",
  "required": {"actions": ["release-notes"]}
}`
	if err := os.WriteFile(filepath.Join(dir, ".keystone/policies/tacoda-org/keystone-policy.json"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	cfg := &config.ProjectConfig{
		Version: config.SchemaVersion,
		Policies: []config.PolicyNode{
			{Name: "tacoda-org", Source: "tacoda/tacoda-org", Version: "v1.0.0"},
		},
	}
	res, err := Verify(dir, cfg, nil)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if res.HasGaps() {
		t.Errorf("expected no gaps when project ships the required item, got %+v", res.RequiredGaps)
	}
}

func TestVerify_RequiredGap_Missing(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".keystone/policies/tacoda-org/dummy.md")
	manifest := `{
  "name": "tacoda-org",
  "version": "1.0.0",
  "required": {"actions": ["release-notes"]}
}`
	if err := os.WriteFile(filepath.Join(dir, ".keystone/policies/tacoda-org/keystone-policy.json"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	cfg := &config.ProjectConfig{
		Version: config.SchemaVersion,
		Policies: []config.PolicyNode{
			{Name: "tacoda-org", Source: "tacoda/tacoda-org", Version: "v1.0.0"},
		},
	}
	res, err := Verify(dir, cfg, nil)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !res.HasGaps() {
		t.Fatalf("expected gap when no layer satisfies required, got none")
	}
	if res.HasErrors() {
		t.Errorf("required gaps should be advisory, not errors: %+v", res.Violations)
	}
	if len(res.RequiredGaps) != 1 {
		t.Fatalf("RequiredGaps = %d, want 1", len(res.RequiredGaps))
	}
	g := res.RequiredGaps[0]
	if g.Policy != "tacoda-org" || g.Port != "actions" || g.Item != "release-notes" {
		t.Errorf("RequiredGap = %+v, want {tacoda-org actions release-notes}", g)
	}
}

func TestVerify_RequiredGap_OuterPolicySatisfies(t *testing.T) {
	dir := t.TempDir()
	// Outer policy ships the required item; inner declares it as required.
	writeFile(t, dir, ".keystone/policies/outer/actions/release-notes.md")
	writeFile(t, dir, ".keystone/policies/inner/dummy.md")
	innerManifest := `{
  "name": "inner",
  "version": "1.0.0",
  "required": {"actions": ["release-notes"]}
}`
	if err := os.WriteFile(filepath.Join(dir, ".keystone/policies/inner/keystone-policy.json"), []byte(innerManifest), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	cfg := &config.ProjectConfig{
		Version: config.SchemaVersion,
		Policies: []config.PolicyNode{
			{
				Name: "outer", Source: "x/outer", Version: "v1",
				Children: []config.PolicyNode{{Name: "inner", Source: "x/inner", Version: "v1"}},
			},
		},
	}
	res, err := Verify(dir, cfg, nil)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if res.HasGaps() {
		t.Errorf("expected no gaps (outer ancestor satisfies inner's required), got %+v", res.RequiredGaps)
	}
}

func TestVerify_NoStrictItems(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.ProjectConfig{
		Version: config.SchemaVersion,
		Policies: []config.PolicyNode{
			{Name: "acme", Source: "acme/policies", Version: "v1"},
		},
	}
	res, err := Verify(dir, cfg, nil)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if res.HasErrors() || res.HasDrift() {
		t.Errorf("expected clean result, got %+v", res)
	}
}

// TestVerify_StrictLocksOn5Layout — a strict item declared with a 4.x key
// (guides, playbooks) or a 5.0 key (rules, hooks) still catches a project
// file that shadows it in the 5.0 layout. Before the fix, a "guides" lock
// looked in .agents/guides/, which no 5.0 project has, and never fired.
func TestVerify_StrictLocksOn5Layout(t *testing.T) {
	cases := []struct {
		key, item, shadow string
	}{
		{"guides", "data-handling", ".agents/rules/process/data-handling.md"},
		{"rules", "data-handling", ".agents/rules/data-handling.md"},
		{"sensors", "secret-scan", ".agents/hooks/secret-scan.md"},
		{"playbooks", "release", ".agents/skills/release/SKILL.md"},
	}
	for _, c := range cases {
		dir := t.TempDir()
		writeFile(t, dir, c.shadow)
		cfg := &config.ProjectConfig{Version: config.SchemaVersion, Policies: []config.PolicyNode{
			{Name: "acme", Source: "acme/policies", Version: "v1", Strict: map[string][]string{c.key: {c.item}}},
		}}
		res, err := Verify(dir, cfg, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(res.Violations) != 1 || res.Violations[0].ShadowPaths[0] != c.shadow {
			t.Errorf("%s/%s: want a violation for %s, got %+v", c.key, c.item, c.shadow, res.Violations)
		}
	}
}

// TestVerify_5PolicyShape — a policy vendored in the 5.0 shape (hooks/,
// rules/) is held to the same depth gate and satisfies `required` the same
// way as a 4.x one.
func TestVerify_5PolicyShape(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".keystone/policies/acme-platform/hooks/rubocop.md")
	cfg := &config.ProjectConfig{Version: config.SchemaVersion, Policies: []config.PolicyNode{
		{Name: "acme-org", Source: "acme/org", Version: "v1", Children: []config.PolicyNode{
			{Name: "acme-platform", Source: "acme/platform", Version: "v1"},
		}},
	}}
	exp := map[string]map[string]string{"acme-platform": {".keystone/policies/acme-platform/hooks/rubocop.md": ""}}
	res, err := Verify(dir, cfg, exp)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.DepthViolations) != 1 || len(res.DepthViolations[0].VendoredSensors) != 1 {
		t.Errorf("a nested policy shipping hooks/ must hit the depth gate: %+v", res.DepthViolations)
	}
	writeFile(t, dir, ".keystone/policies/acme-org/rules/data-handling.md")
	if !requiredSatisfied(dir, ".agents", []string{"acme-org"}, "guides", "data-handling") {
		t.Error("an ancestor's 5.0 rules/ must satisfy a 4.x `guides` requirement")
	}
}
