package claudecode

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// hookOn wraps a single kind:hook bound to a host-phase (or framework) event.
func hookOn(id, event string) primitive.Primitive {
	return primitive.Primitive{
		Frontmatter: primitive.Frontmatter{Kind: "hook", ID: id, Event: event},
	}
}

func TestProjectHooks_BridgePerHostPhase(t *testing.T) {
	root := t.TempDir()
	res, err := ProjectHooks(root, []primitive.Primitive{hookOn("secret-scan", "PreToolUse")})
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
		`"command": "keystone hook fire PreToolUse"`,
		`"statusMessage": "keystone:bridge:PreToolUse"`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("settings missing %q\n%s", want, s)
		}
	}
}

func TestProjectHooks_DistinctPhasesDeduped(t *testing.T) {
	root := t.TempDir()
	// Two hooks on the same phase → ONE bridge entry (keystone dispatches both).
	res, err := ProjectHooks(root, []primitive.Primitive{
		hookOn("a", "PreToolUse"),
		hookOn("b", "PreToolUse"),
		hookOn("c", "Stop"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Added != 2 {
		t.Errorf("expected 2 bridges (PreToolUse, Stop), got %d", res.Added)
	}
	data, _ := os.ReadFile(filepath.Join(root, SettingsRelPath))
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}
	pre := parsed["hooks"].(map[string]any)["PreToolUse"].([]any)
	if len(pre) != 1 {
		t.Errorf("expected 1 bridge group under PreToolUse, got %d", len(pre))
	}
}

func TestProjectHooks_FrameworkEventNotBridged(t *testing.T) {
	root := t.TempDir()
	// pre-verify is a framework event — keystone-fired, never bridged.
	res, err := ProjectHooks(root, []primitive.Primitive{hookOn("review", "pre-verify")})
	if err != nil {
		t.Fatal(err)
	}
	if res.Added != 0 {
		t.Errorf("framework event should not be bridged, got %d", res.Added)
	}
}

func TestProjectHooks_IsIdempotent(t *testing.T) {
	root := t.TempDir()
	p := []primitive.Primitive{hookOn("secret-scan", "PreToolUse"), hookOn("build", "Stop")}
	if _, err := ProjectHooks(root, p); err != nil {
		t.Fatal(err)
	}
	res2, err := ProjectHooks(root, p)
	if err != nil {
		t.Fatal(err)
	}
	if res2.Wrote {
		t.Errorf("second run rewrote settings; expected idempotent no-op (%+v)", res2)
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
	if _, err := ProjectHooks(root, []primitive.Primitive{hookOn("secret-scan", "PreToolUse")}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(settingsPath)
	s := string(data)
	for _, want := range []string{
		`"permissions"`,
		`"echo hi"`,
		`"user-thing"`,
		`"keystone:bridge:PreToolUse"`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("merged settings missing %q\n%s", want, s)
		}
	}
}

func TestProjectHooks_RemovesStaleManagedOnRerun(t *testing.T) {
	root := t.TempDir()
	round1 := []primitive.Primitive{hookOn("a", "PreToolUse"), hookOn("b", "PostToolUse")}
	if _, err := ProjectHooks(root, round1); err != nil {
		t.Fatal(err)
	}
	// Drop the PostToolUse hook — its bridge must be removed.
	round2 := []primitive.Primitive{hookOn("a", "PreToolUse")}
	res, err := ProjectHooks(root, round2)
	if err != nil {
		t.Fatal(err)
	}
	if res.Removed != 2 {
		t.Errorf("expected 2 removed (both prior bridges), got %d", res.Removed)
	}
	data, _ := os.ReadFile(filepath.Join(root, SettingsRelPath))
	s := string(data)
	if strings.Contains(s, "PostToolUse") {
		t.Errorf("stale PostToolUse bridge not removed:\n%s", s)
	}
	if !strings.Contains(s, "keystone:bridge:PreToolUse") {
		t.Errorf("PreToolUse bridge missing:\n%s", s)
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
      { "matcher": "", "hooks": [{ "type": "command", "command": "x", "statusMessage": "keystone:bridge:PreToolUse" }] }
    ]
  }
}
`
	if err := os.WriteFile(settingsPath, []byte(seed), 0o644); err != nil {
		t.Fatal(err)
	}
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

func TestProjectHooks_IgnoresNonHookPrimitives(t *testing.T) {
	root := t.TempDir()
	p := []primitive.Primitive{
		// A guide with an event somehow set → ignored (not kind:hook).
		{Frontmatter: primitive.Frontmatter{Kind: "guide", ID: "g1", Event: "Stop"}},
		// A hook with no event → contributes nothing.
		hookOn("idle", ""),
		// A real host-phase hook → bridged.
		hookOn("real", "Stop"),
	}
	res, err := ProjectHooks(root, p)
	if err != nil {
		t.Fatal(err)
	}
	if res.Added != 1 {
		t.Errorf("expected 1 bridge (only the host-phase hook), got %d", res.Added)
	}
}
