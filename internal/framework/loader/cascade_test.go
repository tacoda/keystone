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
	writeFile(t, dir, "harness/guides/spec.md") // project file, unrelated

	cfg := &config.ProjectConfig{
		Version:     config.SchemaVersion,
		HarnessRoot: "harness",
		Plugins: []config.PluginNode{
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
	writeFile(t, dir, "harness/guides/data-handling.md") // project shadow

	cfg := &config.ProjectConfig{
		Version: config.SchemaVersion,
		Plugins: []config.PluginNode{
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
	if v.Plugin != "acme" || v.Port != "guides" || v.Item != "data-handling" {
		t.Errorf("unexpected violation: %+v", v)
	}
	if v.PathContext != "acme" {
		t.Errorf("PathContext = %q, want %q", v.PathContext, "acme")
	}
	if len(v.ShadowPaths) != 1 || v.ShadowPaths[0] != "harness/guides/data-handling.md" {
		t.Errorf("unexpected shadow paths: %+v", v.ShadowPaths)
	}
}

func TestVerify_NestedPluginPathContext(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "harness/sensors/rubocop.md")

	cfg := &config.ProjectConfig{
		Version: config.SchemaVersion,
		Plugins: []config.PluginNode{
			{
				Name:    "acme-org",
				Source:  "acme/org",
				Version: "v1",
				Children: []config.PluginNode{
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

func TestVerify_NoStrictItems(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.ProjectConfig{
		Version: config.SchemaVersion,
		Plugins: []config.PluginNode{
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
