package claudecode

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

func postureFixture() []primitive.Primitive {
	return []primitive.Primitive{
		{Frontmatter: primitive.Frontmatter{
			Kind: "posture", ID: "default", Description: "x",
			Allow: []string{"Bash(go test:*)", "Read(*)"},
			Ask:   []string{"Bash(git push:*)"},
			Deny:  []string{"Read(.env)"},
		}},
	}
}

// readPermissions pulls settings.permissions[key] as a []string for assertions.
func readPermissions(t *testing.T, abs, key string) []string {
	t.Helper()
	data, err := os.ReadFile(abs)
	if err != nil {
		t.Fatalf("read settings: %v", err)
	}
	var m struct {
		Permissions map[string][]string `json:"permissions"`
	}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("parse settings: %v", err)
	}
	return m.Permissions[key]
}

func contains(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}

func TestProjectPosture_MergesPermissions(t *testing.T) {
	root := t.TempDir()
	if _, err := ProjectPosture(root, postureFixture()); err != nil {
		t.Fatalf("ProjectPosture: %v", err)
	}
	abs := filepath.Join(root, SettingsRelPath)
	if !contains(readPermissions(t, abs, "allow"), "Bash(go test:*)") {
		t.Errorf("allow missing the posture entry: %v", readPermissions(t, abs, "allow"))
	}
	if !contains(readPermissions(t, abs, "ask"), "Bash(git push:*)") {
		t.Errorf("ask missing the posture entry")
	}
	if !contains(readPermissions(t, abs, "deny"), "Read(.env)") {
		t.Errorf("deny missing the posture entry")
	}
}

func TestProjectPosture_PreservesExisting(t *testing.T) {
	root := t.TempDir()
	abs := filepath.Join(root, SettingsRelPath)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	// A user-authored permission the projection must not clobber.
	seed := `{"permissions":{"allow":["Bash(ls:*)"]}}`
	if err := os.WriteFile(abs, []byte(seed), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ProjectPosture(root, postureFixture()); err != nil {
		t.Fatalf("ProjectPosture: %v", err)
	}
	allow := readPermissions(t, abs, "allow")
	if !contains(allow, "Bash(ls:*)") {
		t.Errorf("user permission dropped: %v", allow)
	}
	if !contains(allow, "Bash(go test:*)") {
		t.Errorf("posture permission not merged: %v", allow)
	}
}

func TestProjectPosture_Idempotent(t *testing.T) {
	root := t.TempDir()
	if _, err := ProjectPosture(root, postureFixture()); err != nil {
		t.Fatal(err)
	}
	res, err := ProjectPosture(root, postureFixture())
	if err != nil {
		t.Fatal(err)
	}
	if res.Wrote {
		t.Errorf("second run rewrote settings; expected idempotent no-op")
	}
}
