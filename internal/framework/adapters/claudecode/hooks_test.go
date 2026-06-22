package claudecode

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// sensorWith wraps a single sensor's id + host_triggers into the
// primitive slice the adapter expects. Keeps each test compact.
func sensorWith(id string, triggers ...primitive.HostTrigger) primitive.Primitive {
	return primitive.Primitive{
		Frontmatter: primitive.Frontmatter{
			Kind:         "sensor",
			ID:           id,
			HostTriggers: triggers,
		},
	}
}

// sensorWithSeverity is sensorWith plus a severity tier — used by the
// severity-wrap tests.
func sensorWithSeverity(id, severity string, triggers ...primitive.HostTrigger) primitive.Primitive {
	p := sensorWith(id, triggers...)
	p.Severity = severity
	return p
}

func TestProjectHooks_FirstRunCreatesFile(t *testing.T) {
	root := t.TempDir()
	p := []primitive.Primitive{
		sensorWith("secret-scan",
			primitive.HostTrigger{Phase: "PreToolUse", Matcher: "Edit|Write|MultiEdit",
				Command: "keystone verify --sensor secret-scan", Timeout: 5}),
	}
	res, err := ProjectHooks(root, p)
	if err != nil {
		t.Fatalf("ProjectHooks: %v", err)
	}
	if !res.Wrote || res.Added != 1 || res.Removed != 0 {
		t.Errorf("unexpected result: %+v", res)
	}
	data, err := os.ReadFile(filepath.Join(root, SettingsRelPath))
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	for _, want := range []string{
		`"PreToolUse"`,
		`"matcher": "Edit|Write|MultiEdit"`,
		`"command": "keystone verify --sensor secret-scan"`,
		`"statusMessage": "keystone:secret-scan"`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("settings missing %q\n%s", want, s)
		}
	}
}

func TestProjectHooks_IsIdempotent(t *testing.T) {
	root := t.TempDir()
	p := []primitive.Primitive{
		sensorWith("secret-scan",
			primitive.HostTrigger{Phase: "PreToolUse", Matcher: "Edit|Write",
				Command: "keystone verify --sensor secret-scan", Timeout: 5}),
		sensorWith("build",
			primitive.HostTrigger{Phase: "Stop",
				Command: "go build ./...", Timeout: 60}),
	}
	if _, err := ProjectHooks(root, p); err != nil {
		t.Fatal(err)
	}
	res2, err := ProjectHooks(root, p)
	if err != nil {
		t.Fatal(err)
	}
	if res2.Wrote {
		t.Errorf("second run rewrote settings; expected idempotent no-op (added=%d removed=%d)", res2.Added, res2.Removed)
	}
}

func TestProjectHooks_PreservesUserEntries(t *testing.T) {
	root := t.TempDir()
	settingsPath := filepath.Join(root, SettingsRelPath)
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatal(err)
	}
	seed := `{
  "permissions": { "allow": ["Read"] },
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          { "type": "command", "shell": "bash", "command": "echo hi", "timeout": 1, "statusMessage": "user-thing" }
        ]
      }
    ]
  }
}
`
	if err := os.WriteFile(settingsPath, []byte(seed), 0o644); err != nil {
		t.Fatal(err)
	}
	p := []primitive.Primitive{
		sensorWith("secret-scan",
			primitive.HostTrigger{Phase: "PreToolUse", Matcher: "Edit|Write",
				Command: "keystone verify --sensor secret-scan", Timeout: 5}),
	}
	if _, err := ProjectHooks(root, p); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(settingsPath)
	s := string(data)
	for _, want := range []string{
		`"permissions"`,
		`"echo hi"`,
		`"user-thing"`,
		`"keystone:secret-scan"`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("merged settings missing %q\n%s", want, s)
		}
	}
}

func TestProjectHooks_RemovesPreviouslyManagedEntriesOnRerun(t *testing.T) {
	root := t.TempDir()
	round1 := []primitive.Primitive{
		sensorWith("secret-scan", primitive.HostTrigger{Phase: "PreToolUse", Matcher: "Edit|Write",
			Command: "keystone verify --sensor secret-scan", Timeout: 5}),
		sensorWith("drift", primitive.HostTrigger{Phase: "PostToolUse", Matcher: "Edit|Write",
			Command: "keystone verify --sensor drift", Timeout: 5}),
	}
	if _, err := ProjectHooks(root, round1); err != nil {
		t.Fatal(err)
	}
	// User changed their mind — drop drift, keep secret-scan, add commit-message.
	round2 := []primitive.Primitive{
		sensorWith("secret-scan", primitive.HostTrigger{Phase: "PreToolUse", Matcher: "Edit|Write",
			Command: "keystone verify --sensor secret-scan", Timeout: 5}),
		sensorWith("commit-message", primitive.HostTrigger{Phase: "PreToolUse", Matcher: "Bash",
			Command: "keystone verify --sensor commit-message", Timeout: 5}),
	}
	res, err := ProjectHooks(root, round2)
	if err != nil {
		t.Fatal(err)
	}
	if res.Removed != 2 {
		t.Errorf("expected 2 removals (both prior keystone:* entries), got %d", res.Removed)
	}
	if res.Added != 2 {
		t.Errorf("expected 2 adds, got %d", res.Added)
	}
	data, _ := os.ReadFile(filepath.Join(root, SettingsRelPath))
	s := string(data)
	if strings.Contains(s, `"keystone:drift"`) {
		t.Errorf("expected drift hook removed; settings:\n%s", s)
	}
	if !strings.Contains(s, `"keystone:secret-scan"`) {
		t.Errorf("expected secret-scan hook present; settings:\n%s", s)
	}
	if !strings.Contains(s, `"keystone:commit-message"`) {
		t.Errorf("expected commit-message hook present; settings:\n%s", s)
	}
}

func TestProjectHooks_GroupsByPhaseAndMatcher(t *testing.T) {
	root := t.TempDir()
	p := []primitive.Primitive{
		sensorWith("a", primitive.HostTrigger{Phase: "PreToolUse", Matcher: "Edit|Write", Command: "ka", Timeout: 5}),
		sensorWith("b", primitive.HostTrigger{Phase: "PreToolUse", Matcher: "Edit|Write", Command: "kb", Timeout: 5}),
		sensorWith("c", primitive.HostTrigger{Phase: "PreToolUse", Matcher: "Bash", Command: "kc", Timeout: 5}),
	}
	if _, err := ProjectHooks(root, p); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(root, SettingsRelPath))
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}
	hooksMap := parsed["hooks"].(map[string]any)
	pre := hooksMap["PreToolUse"].([]any)
	if len(pre) != 2 {
		t.Fatalf("expected 2 matcher groups under PreToolUse, got %d: %v", len(pre), pre)
	}
}

func TestProjectHooks_DropsHooksKeyWhenEmpty(t *testing.T) {
	root := t.TempDir()
	settingsPath := filepath.Join(root, SettingsRelPath)
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatal(err)
	}
	seed := `{
  "permissions": { "allow": [] },
  "hooks": {
    "PreToolUse": [
      { "matcher": "Edit", "hooks": [{ "type": "command", "command": "x", "statusMessage": "keystone:old" }] }
    ]
  }
}
`
	if err := os.WriteFile(settingsPath, []byte(seed), 0o644); err != nil {
		t.Fatal(err)
	}
	// Re-run with empty sensor set — managed entries are removed; no replacements.
	if _, err := ProjectHooks(root, nil); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(settingsPath)
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}
	if _, exists := parsed["hooks"]; exists {
		t.Errorf("expected hooks key removed when empty; got %v", parsed["hooks"])
	}
	if _, exists := parsed["permissions"]; !exists {
		t.Errorf("expected permissions key preserved")
	}
}

func TestProjectHooks_IgnoresNonSensorPrimitives(t *testing.T) {
	root := t.TempDir()
	p := []primitive.Primitive{
		// A guide with HostTriggers set (somehow) should be ignored.
		{Frontmatter: primitive.Frontmatter{Kind: "guide", ID: "g1",
			HostTriggers: []primitive.HostTrigger{{Phase: "Stop", Command: "x"}}}},
		// A sensor with no triggers contributes nothing.
		sensorWith("idle"),
		// A real sensor with a trigger gets projected.
		sensorWith("real", primitive.HostTrigger{Phase: "Stop", Command: "go test", Timeout: 30}),
	}
	res, err := ProjectHooks(root, p)
	if err != nil {
		t.Fatal(err)
	}
	if res.Added != 1 {
		t.Errorf("expected 1 hook (only the sensor trigger), got %d", res.Added)
	}
}

func TestProjectHooks_SeverityMust_Unwrapped(t *testing.T) {
	root := t.TempDir()
	p := []primitive.Primitive{
		sensorWithSeverity("strict", "must",
			primitive.HostTrigger{Phase: "PreToolUse", Matcher: "Edit", Command: "keystone verify --sensor strict", Timeout: 5}),
	}
	if _, err := ProjectHooks(root, p); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(root, SettingsRelPath))
	s := string(data)
	// `must` (or default) leaves the command intact — no wrapper.
	if !strings.Contains(s, `"command": "keystone verify --sensor strict"`) {
		t.Errorf("must severity should leave command unwrapped, got:\n%s", s)
	}
}

func TestProjectHooks_SeverityShould_WrapsWarning(t *testing.T) {
	root := t.TempDir()
	p := []primitive.Primitive{
		sensorWithSeverity("soft", "should",
			primitive.HostTrigger{Phase: "Stop", Command: "go vet ./...", Timeout: 60}),
	}
	if _, err := ProjectHooks(root, p); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(root, SettingsRelPath))
	s := string(data)
	if !strings.Contains(s, `( go vet ./... )`) || !strings.Contains(s, `non-blocking warning`) {
		t.Errorf("should severity didn't wrap with warning fallback:\n%s", s)
	}
	if !strings.Contains(s, "keystone:soft") {
		t.Errorf("statusMessage missing:\n%s", s)
	}
}

func TestProjectHooks_SeverityMay_WrapsSilent(t *testing.T) {
	root := t.TempDir()
	p := []primitive.Primitive{
		sensorWithSeverity("info", "may",
			primitive.HostTrigger{Phase: "Stop", Command: "go test -short ./...", Timeout: 60}),
	}
	if _, err := ProjectHooks(root, p); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(root, SettingsRelPath))
	s := string(data)
	if !strings.Contains(s, `( go test -short ./... ) >/dev/null 2>&1 || true`) {
		t.Errorf("may severity didn't wrap silently:\n%s", s)
	}
}

func TestProjectHooks_SeverityDefaultsToMust(t *testing.T) {
	root := t.TempDir()
	p := []primitive.Primitive{
		sensorWithSeverity("unset", "",
			primitive.HostTrigger{Phase: "PreToolUse", Matcher: "Edit", Command: "cmd", Timeout: 5}),
	}
	if _, err := ProjectHooks(root, p); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(root, SettingsRelPath))
	s := string(data)
	if !strings.Contains(s, `"command": "cmd"`) {
		t.Errorf("empty severity should behave as must (unwrapped), got:\n%s", s)
	}
}
