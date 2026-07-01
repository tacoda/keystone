package migrations

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// seed30Install lays down a 3.0-layout install: a .harness/ root with a
// primitive and the vestigial context.json.
func seed30Install(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	files := map[string]string{
		".harness/guides/idioms/go/stdlib-first.md": "---\nkind: guide\nid: guides/idioms/go/stdlib-first\ndescription: x\nglobs:\n  - \"**/*.go\"\n---\n# Stdlib first\n",
		".harness/hooks/build.md":                   "---\nkind: hook\nid: build\ndescription: x\nmode: computational\nevent: Stop\nrun: go build ./...\n---\nbody\n",
		".harness/context.json":                     "{}\n",
	}
	for rel, body := range files {
		abs := filepath.Join(tmp, rel)
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return tmp
}

func runV4(t *testing.T, tmp string, plan func(string) (*Plan, error)) {
	t.Helper()
	p, err := plan(tmp)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if err := p.Execute(tmp); err != nil {
		t.Fatalf("execute: %v", err)
	}
}

func TestV4_Up_RenamesRootDropsContextRebuildsIndex(t *testing.T) {
	tmp := seed30Install(t)
	runV4(t, tmp, planUp_4_0)

	// .harness/ is gone; .charter/ holds the primitive.
	if dirExists(filepath.Join(tmp, ".harness")) {
		t.Error(".harness/ should be renamed away")
	}
	guide := filepath.Join(tmp, ".charter/guides/idioms/go/stdlib-first.md")
	if _, err := os.Stat(guide); err != nil {
		t.Errorf("guide missing under .charter/: %v", err)
	}
	// context.json dropped.
	if _, err := os.Stat(filepath.Join(tmp, ".charter/context.json")); !os.IsNotExist(err) {
		t.Error("context.json should be removed")
	}
	// INDEX rebuilt at the new root.
	for _, name := range []string{"INDEX.json", "INDEX.lite.json"} {
		if _, err := os.Stat(filepath.Join(tmp, ".charter", name)); err != nil {
			t.Errorf("%s not rebuilt: %v", name, err)
		}
	}
}

func TestV4_Up_FoldsHooksToSensors(t *testing.T) {
	tmp := seed30Install(t)
	runV4(t, tmp, planUp_4_0)

	if dirExists(filepath.Join(tmp, ".charter/hooks")) {
		t.Error("hooks/ should be folded away")
	}
	folded, err := os.ReadFile(filepath.Join(tmp, ".charter/sensors/build.md"))
	if err != nil {
		t.Fatalf("folded sensor missing: %v", err)
	}
	s := string(folded)
	if !strings.Contains(s, "kind: sensor") || strings.Contains(s, "kind: hook") {
		t.Errorf("kind not rewritten to sensor:\n%s", s)
	}
	if !strings.Contains(s, "on: Stop") || strings.Contains(s, "event: Stop") {
		t.Errorf("event: not rewritten to on::\n%s", s)
	}
}

func TestV4_Up_Idempotent(t *testing.T) {
	tmp := seed30Install(t)
	runV4(t, tmp, planUp_4_0)
	// Second Up must not error (root already renamed).
	runV4(t, tmp, planUp_4_0)
	if _, err := os.Stat(filepath.Join(tmp, ".charter/guides/idioms/go/stdlib-first.md")); err != nil {
		t.Errorf("guide missing after second Up: %v", err)
	}
}

func TestV4_Down_ReversesRootRename(t *testing.T) {
	tmp := seed30Install(t)
	runV4(t, tmp, planUp_4_0)
	runV4(t, tmp, planDown_4_0)

	if dirExists(filepath.Join(tmp, ".charter")) {
		t.Error(".charter/ should be renamed back to .harness/")
	}
	if _, err := os.Stat(filepath.Join(tmp, ".harness/guides/idioms/go/stdlib-first.md")); err != nil {
		t.Errorf("guide missing under .harness/ after Down: %v", err)
	}
}

func TestV4_Up_FreshInstallNoop(t *testing.T) {
	tmp := t.TempDir() // no .harness/ and no .charter/
	p, err := planUp_4_0(tmp)
	if err != nil {
		t.Fatalf("planUp_4_0: %v", err)
	}
	if err := p.Execute(tmp); err != nil {
		t.Fatalf("execute: %v", err)
	}
}
