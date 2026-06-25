package migrations

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompareVersion(t *testing.T) {
	cases := []struct {
		a, b string
		want int // -1 / 0 / +1 (sign only)
	}{
		{"2.0", "2.0", 0},
		{"2.0", "2.0.1", -1},
		{"2.0", "2.1", -1},
		{"2.1", "2.10", -1},
		{"2.1", "2.0", +1},
		{"3.0", "2.99", +1},
		{"2.0", "2", 0}, // missing segment treated as zero
	}
	for _, c := range cases {
		got := CompareVersion(c.a, c.b)
		gotSign := sign(got)
		if gotSign != c.want {
			t.Errorf("CompareVersion(%q, %q) = %d (sign %d), want sign %d", c.a, c.b, got, gotSign, c.want)
		}
	}
}

func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return +1
	default:
		return 0
	}
}

func TestPending(t *testing.T) {
	all := All()
	if len(all) == 0 {
		t.Fatal("registry is empty — package init didn't register migrations")
	}

	cases := []struct {
		name    string
		applied []string
		wantLen int
	}{
		{"empty applied → all pending", nil, len(all)},
		{"applied latest → none pending", []string{all[len(all)-1].Version}, 0},
		{"applied first → all but first pending", []string{all[0].Version}, len(all) - 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Pending(c.applied)
			if len(got) != c.wantLen {
				t.Errorf("Pending(%v) returned %d, want %d", c.applied, len(got), c.wantLen)
			}
		})
	}
}

func TestPlanExecute_StopsOnError(t *testing.T) {
	calls := 0
	p := &Plan{}
	p.Add("first", func(_ string) error { calls++; return nil })
	p.Add("second (fails)", func(_ string) error { calls++; return os.ErrPermission })
	p.Add("third (should not run)", func(_ string) error { calls++; return nil })

	if err := p.Execute("/dev/null"); err == nil {
		t.Fatal("expected error from failing step")
	}
	if calls != 2 {
		t.Errorf("calls = %d, want 2 (third should be skipped)", calls)
	}
}

// TestV2_0_Roundtrip builds a synthetic 1.x install, runs the 2.0 Up,
// asserts the 2.0 shape on disk, then runs Down and asserts the
// pre-migration shape is restored. Confirms paired Up/Down don't drift.
func TestV2_0_Roundtrip(t *testing.T) {
	dir := t.TempDir()

	// Pre-migration shape.
	mustMkdir(t, filepath.Join(dir, "harness"))
	mustWrite(t, filepath.Join(dir, "harness", "keystone.lock.json"), `{"version":1,"keystone":{"version":"2.0.3"},"policies":{}}`)
	mustMkdir(t, filepath.Join(dir, "harness", "plugins", "example"))
	mustWrite(t, filepath.Join(dir, "harness", "plugins", "example", "keystone-plugin.json"), `{"name":"example","version":"0.1.0"}`)
	mustWrite(t, filepath.Join(dir, "keystone.json"), `{"version":"1","plugins":[{"name":"example","source":"tacoda/example","version":"0.1.0"}]}`)

	v2_0 := Find("2.0")
	if v2_0 == nil {
		t.Fatal("2.0 migration not registered")
	}

	// Up.
	up, err := v2_0.Up(dir)
	if err != nil {
		t.Fatalf("plan up: %v", err)
	}
	if err := up.Execute(dir); err != nil {
		t.Fatalf("execute up: %v", err)
	}
	assertExists(t, filepath.Join(dir, ".keystone", "harness"))
	assertNotExists(t, filepath.Join(dir, "harness"))
	assertExists(t, filepath.Join(dir, ".keystone", "lockfile.json"))
	assertExists(t, filepath.Join(dir, ".keystone", "harness", "policies", "example", "keystone-policy.json"))
	assertNotExists(t, filepath.Join(dir, ".keystone", "harness", "plugins"))
	assertCfgField(t, filepath.Join(dir, "keystone.json"), "version", "2")
	assertCfgFieldExists(t, filepath.Join(dir, "keystone.json"), "policies")
	assertCfgFieldAbsent(t, filepath.Join(dir, "keystone.json"), "plugins")

	// Down.
	down, err := v2_0.Down(dir)
	if err != nil {
		t.Fatalf("plan down: %v", err)
	}
	if err := down.Execute(dir); err != nil {
		t.Fatalf("execute down: %v", err)
	}
	assertExists(t, filepath.Join(dir, "harness"))
	assertNotExists(t, filepath.Join(dir, ".keystone", "harness"))
	assertExists(t, filepath.Join(dir, "harness", "keystone.lock.json"))
	assertExists(t, filepath.Join(dir, "harness", "plugins", "example", "keystone-plugin.json"))
	assertNotExists(t, filepath.Join(dir, "harness", "policies"))
	assertCfgField(t, filepath.Join(dir, "keystone.json"), "version", "1")
	assertCfgFieldExists(t, filepath.Join(dir, "keystone.json"), "plugins")
	assertCfgFieldAbsent(t, filepath.Join(dir, "keystone.json"), "policies")
}

// TestV2_0_Up_RefusesAmbiguousState confirms a project with both legacy
// and 2.0 layouts side-by-side surfaces the conflict instead of
// silently picking one.
func TestV2_0_Up_RefusesAmbiguousState(t *testing.T) {
	dir := t.TempDir()
	mustMkdir(t, filepath.Join(dir, "harness"))
	mustMkdir(t, filepath.Join(dir, ".keystone", "harness"))

	v2_0 := Find("2.0")
	if v2_0 == nil {
		t.Fatal("2.0 migration not registered")
	}
	if _, err := v2_0.Up(dir); err == nil {
		t.Fatal("expected error when both legacy and target harness dirs exist")
	}
}

// TestV2_1_Roundtrip writes a lockfile with the pre-2.1 plugin_version
// tag, runs 2.1 Up, asserts the field was renamed, then runs Down and
// asserts the legacy tag is restored.
func TestV2_1_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	mustMkdir(t, filepath.Join(dir, ".keystone"))
	lockPath := filepath.Join(dir, ".keystone", "lockfile.json")
	mustWrite(t, lockPath, `{
  "version": 1,
  "keystone": {"version": "2.0.3"},
  "policies": {
    "example": {
      "source_ref": "tacoda/example",
      "resolved_sha": "abc123",
      "plugin_version": "0.1.0",
      "version": "v0.1.0",
      "files": {}
    }
  }
}`)

	v2_1 := Find("2.1")
	if v2_1 == nil {
		t.Fatal("2.1 migration not registered")
	}

	up, err := v2_1.Up(dir)
	if err != nil {
		t.Fatalf("plan up: %v", err)
	}
	if err := up.Execute(dir); err != nil {
		t.Fatalf("execute up: %v", err)
	}
	got := readLockfileRaw(t, lockPath)
	entry := got["policies"].(map[string]any)["example"].(map[string]any)
	if _, has := entry["policy_version"]; !has {
		t.Errorf("after up: policy_version missing; entry = %+v", entry)
	}
	if _, has := entry["plugin_version"]; has {
		t.Errorf("after up: plugin_version should be removed; entry = %+v", entry)
	}

	down, err := v2_1.Down(dir)
	if err != nil {
		t.Fatalf("plan down: %v", err)
	}
	if err := down.Execute(dir); err != nil {
		t.Fatalf("execute down: %v", err)
	}
	got = readLockfileRaw(t, lockPath)
	entry = got["policies"].(map[string]any)["example"].(map[string]any)
	if _, has := entry["plugin_version"]; !has {
		t.Errorf("after down: plugin_version missing; entry = %+v", entry)
	}
	if _, has := entry["policy_version"]; has {
		t.Errorf("after down: policy_version should be removed; entry = %+v", entry)
	}
}

// TestV2_1_Up_Idempotent confirms re-running 2.1 Up on already-migrated
// content is a safe no-op (no file written, no policy_version clobbered).
func TestV2_1_Up_Idempotent(t *testing.T) {
	dir := t.TempDir()
	mustMkdir(t, filepath.Join(dir, ".keystone"))
	lockPath := filepath.Join(dir, ".keystone", "lockfile.json")
	original := `{"version":1,"keystone":{"version":"2.1.0"},"policies":{"example":{"source_ref":"tacoda/example","resolved_sha":"abc","policy_version":"0.1.0","version":"v0.1.0","files":{}}}}`
	mustWrite(t, lockPath, original)

	v2_1 := Find("2.1")
	up, _ := v2_1.Up(dir)
	if err := up.Execute(dir); err != nil {
		t.Fatalf("execute up: %v", err)
	}
	got := readLockfileRaw(t, lockPath)
	entry := got["policies"].(map[string]any)["example"].(map[string]any)
	if entry["policy_version"] != "0.1.0" {
		t.Errorf("idempotent up clobbered policy_version: %+v", entry)
	}
}

// helpers

func mustMkdir(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", p, err)
	}
}

func mustWrite(t *testing.T, p, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("mkdir parent of %s: %v", p, err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
}

func assertExists(t *testing.T, p string) {
	t.Helper()
	if _, err := os.Stat(p); err != nil {
		t.Errorf("expected to exist: %s (%v)", p, err)
	}
}

func assertNotExists(t *testing.T, p string) {
	t.Helper()
	if _, err := os.Stat(p); err == nil {
		t.Errorf("expected to NOT exist: %s", p)
	}
}

func assertCfgField(t *testing.T, path, key, want string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	got, _ := doc[key].(string)
	if got != want {
		t.Errorf("%s[%q] = %q, want %q", filepath.Base(path), key, got, want)
	}
}

func assertCfgFieldExists(t *testing.T, path, key string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	if _, ok := doc[key]; !ok {
		t.Errorf("%s: expected field %q to be present", filepath.Base(path), key)
	}
}

func assertCfgFieldAbsent(t *testing.T, path, key string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	if _, ok := doc[key]; ok {
		t.Errorf("%s: expected field %q to be absent", filepath.Base(path), key)
	}
}

func readLockfileRaw(t *testing.T, p string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("parse %s: %v", p, err)
	}
	return out
}

func TestV2_2_Up_OnEmptyProject(t *testing.T) {
	// No .keystone/harness/ → migration is a no-op.
	tmp := t.TempDir()
	plan, err := planUp_2_2(tmp)
	if err != nil {
		t.Fatalf("planUp_2_2: %v", err)
	}
	if len(plan.Steps) != 0 {
		t.Errorf("expected 0 steps on empty project, got %d", len(plan.Steps))
	}
}

func TestV2_2_Up_Idempotent(t *testing.T) {
	tmp := t.TempDir()

	// Seed a minimal harness: one sensor with host_triggers, one guide
	// with globs, plus keystone.json.
	mustMkdir(t, filepath.Join(tmp, ".keystone/harness/sensors"))
	mustMkdir(t, filepath.Join(tmp, ".keystone/harness/guides/idioms/go"))
	mustWrite(t, filepath.Join(tmp, ".keystone/harness/sensors/build.md"),
		"---\nkind: hook\nid: build\ndescription: x\nhost_triggers:\n  - phase: Stop\n    command: go build ./...\n    timeout: 60\n---\nbody\n")
	mustWrite(t, filepath.Join(tmp, ".keystone/harness/guides/idioms/go/stdlib-first.md"),
		"---\nkind: rule\nid: rules/idioms/go/stdlib-first\ndescription: x\nglobs:\n  - \"cmd/**/*.go\"\n---\n# Stdlib first\n\n## IRON LAW\n\n- One.\n")
	mustWrite(t, filepath.Join(tmp, "keystone.json"),
		"{\"version\":\"2\",\"policies\":[]}\n")

	// First Up run.
	plan, err := planUp_2_2(tmp)
	if err != nil {
		t.Fatalf("planUp_2_2: %v", err)
	}
	if err := plan.Execute(tmp); err != nil {
		t.Fatalf("plan.Execute first run: %v", err)
	}
	// Confirm expected artifacts landed.
	for _, want := range []string{
		".keystone/INDEX.json",
		".keystone/INDEX.lite.json",
		".claude/rules/go-stdlib-first.md",
		".claude/settings.json",
		"AGENTS.md",
	} {
		if _, err := os.Stat(filepath.Join(tmp, want)); err != nil {
			t.Errorf("expected %s after migration: %v", want, err)
		}
	}

	// Second run on already-migrated tree — should still complete cleanly.
	plan2, err := planUp_2_2(tmp)
	if err != nil {
		t.Fatalf("planUp_2_2 second: %v", err)
	}
	if err := plan2.Execute(tmp); err != nil {
		t.Fatalf("plan.Execute second run: %v", err)
	}
}


func TestV2_3_Up_OnEmptyProject(t *testing.T) {
	tmp := t.TempDir()
	plan, err := planUp_2_3(tmp)
	if err != nil {
		t.Fatalf("planUp_2_3: %v", err)
	}
	if len(plan.Steps) != 0 {
		t.Errorf("expected 0 steps on empty project, got %d", len(plan.Steps))
	}
}

func TestV2_3_Up_CreatesConcernsDir(t *testing.T) {
	tmp := t.TempDir()

	mustMkdir(t, filepath.Join(tmp, ".keystone/harness/sensors"))
	mustWrite(t, filepath.Join(tmp, ".keystone/harness/sensors/build.md"),
		"---\nkind: hook\nid: build\ndescription: x\n---\nbody\n")
	mustWrite(t, filepath.Join(tmp, "keystone.json"),
		"{\"version\":\"2\",\"policies\":[]}\n")

	plan, err := planUp_2_3(tmp)
	if err != nil {
		t.Fatalf("planUp_2_3: %v", err)
	}
	if err := plan.Execute(tmp); err != nil {
		t.Fatalf("plan.Execute: %v", err)
	}
	if info, err := os.Stat(filepath.Join(tmp, ".keystone/harness/concerns")); err != nil || !info.IsDir() {
		t.Errorf("expected .keystone/harness/concerns/ to exist: err=%v info=%v", err, info)
	}
}

func TestV2_3_Up_Idempotent(t *testing.T) {
	tmp := t.TempDir()

	// Seed: one sensor with host_triggers + severity, one concern,
	// one persona that includes the concern.
	mustMkdir(t, filepath.Join(tmp, ".keystone/harness/sensors"))
	mustMkdir(t, filepath.Join(tmp, ".keystone/harness/concerns"))
	mustMkdir(t, filepath.Join(tmp, ".keystone/harness/personas"))
	mustWrite(t, filepath.Join(tmp, ".keystone/harness/sensors/build.md"),
		"---\nkind: hook\nid: build\ndescription: x\nseverity: should\nhost_triggers:\n  - phase: Stop\n    command: go build ./...\n    timeout: 60\n---\nbody\n")
	mustWrite(t, filepath.Join(tmp, ".keystone/harness/concerns/shared-tools.md"),
		"---\nkind: concern\nid: shared-tools\ndescription: x\ntools:\n  - Read\n  - Grep\ntags:\n  - composition\n---\nbody\n")
	mustWrite(t, filepath.Join(tmp, ".keystone/harness/personas/reviewer.md"),
		"---\nkind: agent\nid: reviewer\ndescription: x\ntools:\n  - Bash\nincludes:\n  - shared-tools\ntags:\n  - review\n---\nbody\n")
	mustWrite(t, filepath.Join(tmp, "keystone.json"),
		"{\"version\":\"2\",\"policies\":[]}\n")

	// First Up.
	plan, err := planUp_2_3(tmp)
	if err != nil {
		t.Fatalf("planUp_2_3: %v", err)
	}
	if err := plan.Execute(tmp); err != nil {
		t.Fatalf("first Execute: %v", err)
	}

	// Second Up — must complete cleanly.
	plan2, err := planUp_2_3(tmp)
	if err != nil {
		t.Fatalf("planUp_2_3 second: %v", err)
	}
	if err := plan2.Execute(tmp); err != nil {
		t.Fatalf("second Execute: %v", err)
	}

	// Verify the settings.json carries the should-wrap.
	data, err := os.ReadFile(filepath.Join(tmp, ".claude/settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "non-blocking warning") {
		t.Errorf(".claude/settings.json missing should-severity wrapper:\n%s", data)
	}
}
