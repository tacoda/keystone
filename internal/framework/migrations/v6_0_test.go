package migrations

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// seed50InstallWithTools lays down a 5.x-layout install that still has a
// .agents/tools/ tree.
func seed50InstallWithTools(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	writeAll(t, tmp, map[string]string{
		".agents/rules/keep.md":         "---\nkind: rule\nid: keep\ndescription: x\n---\n",
		".agents/tools/grep-symbols.md": "---\nkind: tool\nid: grep-symbols\ndescription: x\ntransport: cli\nrun: ./grep.sh\n---\n# grep\n",
		".agents/tools/README.md":       "# Tools\n",
		".keystone/lockfile.json":       "{\"version\":1,\"migrations_applied\":[\"5.0\"]}\n",
	})
	return tmp
}

func TestV6_Up_ArchivesTools(t *testing.T) {
	tmp := seed50InstallWithTools(t)
	m := Find("6.0")
	if m == nil {
		t.Fatal("6.0 migration not registered")
	}
	runV6(t, tmp, m.Up)

	if _, err := os.Stat(filepath.Join(tmp, ".agents/tools/grep-symbols.md")); !os.IsNotExist(err) {
		t.Errorf(".agents/tools/ should hold only the pointer after Up, got err = %v", err)
	}
	mustExist(t, tmp, ".keystone/archive/6.0/tools/grep-symbols.md", ".agents/tools/README.md")
	body, err := os.ReadFile(filepath.Join(tmp, ".agents/tools/README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "retired in 6.0") {
		t.Errorf("tools pointer missing the retirement note: %s", body)
	}
}

func TestV6_Down_RestoresTools(t *testing.T) {
	tmp := seed50InstallWithTools(t)
	m := Find("6.0")
	runV6(t, tmp, m.Up)
	runV6(t, tmp, m.Down)

	mustExist(t, tmp, ".agents/tools/grep-symbols.md")
	if dirExists(filepath.Join(tmp, ".keystone/archive/6.0/tools")) {
		t.Error("archive should be gone after Down")
	}
}

func TestV6_Up_NoToolsIsNoop(t *testing.T) {
	tmp := t.TempDir()
	writeAll(t, tmp, map[string]string{
		".agents/rules/keep.md":   "---\nkind: rule\nid: keep\ndescription: x\n---\n",
		".keystone/lockfile.json": "{\"version\":1,\"migrations_applied\":[\"5.0\"]}\n",
	})
	m := Find("6.0")
	plan, err := m.Up(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Steps) != 0 {
		t.Errorf("expected a no-op plan without tools/, got %d step(s)", len(plan.Steps))
	}
}

func TestV6_Up_RefusesExistingArchive(t *testing.T) {
	tmp := seed50InstallWithTools(t)
	writeAll(t, tmp, map[string]string{".keystone/archive/6.0/tools/x.md": "conflict\n"})
	m := Find("6.0")
	if _, err := m.Up(tmp); err == nil {
		t.Error("expected a conflict error when the archive already exists")
	}
}

func runV6(t *testing.T, tmp string, dir func(string) (*Plan, error)) {
	t.Helper()
	plan, err := dir(tmp)
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range plan.Steps {
		if err := step.Op(tmp); err != nil {
			t.Fatalf("step %q: %v", step.Desc, err)
		}
	}
}
