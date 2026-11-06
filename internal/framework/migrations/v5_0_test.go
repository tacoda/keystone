package migrations

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	guide40    = "---\nkind: guide\nid: guides/idioms/go/stdlib-first\ndescription: x\nglobs:\n  - \"**/*.go\"\n---\n# Stdlib first\n"
	sensor40   = "---\nkind: sensor\nid: build\ndescription: x\nmode: computational\non: Stop\nrun: go build ./...\n---\nbody\n"
	playbook40 = "---\nkind: playbook\nid: task\ndescription: End-to-end task workflow.\n---\n# task\n"
)

// seed40Install lays down a 4.0-layout install under .charter/.
func seed40Install(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	writeAll(t, tmp, map[string]string{
		".charter/guides/idioms/go/stdlib-first.md": guide40,
		".charter/rules/README.md":                  "# Rules\n",
		".charter/corpus/state/CODEBASE_STATE.md":   "---\nkind: corpus\nid: corpus/state/CODEBASE_STATE\ndescription: x\n---\n",
		".charter/sensors/build.md":                 sensor40,
		".charter/playbooks/task.md":                playbook40,
		".charter/playbooks/README.md":              "# Playbooks\n",
		".charter/adapters/codex/activation.md":     "# codex\n",
		".charter/learning/inbox/.gitkeep":          "",
		".charter/policies/acme/guides/x.md":        guide40,
		".charter/lockfile.json":                    "{\"version\":1,\"migrations_applied\":[\"4.0\"]}\n",
		".charter/INDEX.json":                       "{}\n",
		".charter/README.md":                        "# Charter\n",
		".claude/commands/keystone-verify.md":       "4.x projection\n",
		".claude/commands/mine.md":                  "user file\n",
		".gitignore":                                "/dist/\n.charter/policies/\n",
	})
	return tmp
}

func writeAll(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, body := range files {
		abs := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func mustExist(t *testing.T, root string, rels ...string) {
	t.Helper()
	for _, rel := range rels {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Errorf("missing %s: %v", rel, err)
		}
	}
}

func TestV5_Up_MovesLayout(t *testing.T) {
	tmp := seed40Install(t)
	runV4(t, tmp, planUp_5_0)

	if dirExists(filepath.Join(tmp, ".charter")) {
		t.Error(".charter/ should be gone after Up")
	}
	mustExist(t, tmp,
		".agents/rules/idioms/go/stdlib-first.md",
		".agents/rules/README.md",
		".agents/references/state/CODEBASE_STATE.md",
		".agents/hooks/build.md",
		".agents/skills/task/SKILL.md",
		".agents/README.md",
		".keystone/hosts/codex/activation.md",
		".keystone/learning/inbox/.gitkeep",
		".keystone/policies/acme/guides/x.md",
		".keystone/lockfile.json",
		".keystone/INDEX.json",
		".keystone/INDEX.lite.json",
	)
	// Moves only: file content is byte-identical (iron law 3).
	got, _ := os.ReadFile(filepath.Join(tmp, ".agents/skills/task/SKILL.md"))
	if string(got) != playbook40 {
		t.Errorf("folded playbook content changed:\n%s", got)
	}
	// The retired playbooks README is dropped, not folded into a skill.
	if _, err := os.Stat(filepath.Join(tmp, ".agents/skills/README/SKILL.md")); err == nil {
		t.Error("playbooks/README.md must not become a skill")
	}
	gi, _ := os.ReadFile(filepath.Join(tmp, ".gitignore"))
	if !strings.Contains(string(gi), ".keystone/policies/\n") || !strings.Contains(string(gi), ".charter/policies/\n") {
		t.Errorf(".gitignore should keep the old line and add the new one:\n%s", gi)
	}
	// INDEX was rebuilt from .agents/ with normalized kinds.
	idx, _ := os.ReadFile(filepath.Join(tmp, ".keystone/INDEX.json"))
	for _, want := range []string{`"rule"`, `"hook"`, `"skill"`} {
		if !strings.Contains(string(idx), want) {
			t.Errorf("rebuilt INDEX.json missing kind %s", want)
		}
	}
}

// TestV5_Up_RecordsLegacyProjections — 4.x keystone-* projections are
// recorded as owned; user files beside them are not.
func TestV5_Up_RecordsLegacyProjections(t *testing.T) {
	tmp := seed40Install(t)
	runV4(t, tmp, planUp_5_0)
	lf, _ := os.ReadFile(filepath.Join(tmp, ".keystone/lockfile.json"))
	if !strings.Contains(string(lf), ".claude/commands/keystone-verify.md") {
		t.Errorf("4.x projection not recorded:\n%s", lf)
	}
	if strings.Contains(string(lf), "mine.md") {
		t.Errorf("user file must not be claimed:\n%s", lf)
	}
}

// TestV5_Up_RemovesLegacyEntrypoints — 4.x generated CHARTER.md and a
// pointer-only AGENTS.md go; a user-written AGENTS.md stays.
func TestV5_Up_RemovesLegacyEntrypoints(t *testing.T) {
	tmp := seed40Install(t)
	writeAll(t, tmp, map[string]string{
		"CHARTER.md":     "# Charter\n\nThis repo is governed by a **keystone charter** — x\n",
		"CONVENTIONS.md": "# Charter\n\nYou **must** read [`CHARTER.md`](CHARTER.md) now\n",
		"AGENTS.md":      "# Our agents\n\nHand-written.\n",
	})
	runV4(t, tmp, planUp_5_0)
	for _, gone := range []string{"CHARTER.md", "CONVENTIONS.md"} {
		if _, err := os.Stat(filepath.Join(tmp, gone)); !os.IsNotExist(err) {
			t.Errorf("%s should be removed", gone)
		}
	}
	mustExist(t, tmp, "AGENTS.md")
}

func TestV5_Up_IsIdempotent(t *testing.T) {
	tmp := seed40Install(t)
	runV4(t, tmp, planUp_5_0)
	p, err := planUp_5_0(tmp)
	if err != nil {
		t.Fatalf("second plan: %v", err)
	}
	if len(p.Steps) != 0 {
		t.Errorf("second Up should be a no-op, got %d steps", len(p.Steps))
	}
}

// TestV5_Up_MergesExistingAgentsDir — a repo that already keeps Codex
// skills in .agents/skills/ keeps them; the charter merges in beside them.
func TestV5_Up_MergesExistingAgentsDir(t *testing.T) {
	tmp := seed40Install(t)
	writeAll(t, tmp, map[string]string{".agents/skills/mine/SKILL.md": "---\nname: mine\ndescription: x\n---\n"})
	runV4(t, tmp, planUp_5_0)
	mustExist(t, tmp, ".agents/skills/mine/SKILL.md", ".agents/skills/task/SKILL.md")
}

func TestV5_Up_RefusesConflictingFile(t *testing.T) {
	tmp := seed40Install(t)
	writeAll(t, tmp, map[string]string{".agents/skills/task/SKILL.md": "different\n"})
	p, err := planUp_5_0(tmp)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if err := p.Execute(tmp); err == nil || !strings.Contains(err.Error(), "resolve by hand") {
		t.Fatalf("expected a resolve-by-hand conflict, got %v", err)
	}
	got, _ := os.ReadFile(filepath.Join(tmp, ".agents/skills/task/SKILL.md"))
	if string(got) != "different\n" {
		t.Error("the existing file must not be overwritten")
	}
}

func TestV5_Up_RefusesUnknownEntryBeforeMoving(t *testing.T) {
	tmp := seed40Install(t)
	writeAll(t, tmp, map[string]string{".charter/mystery/x.md": "x\n"})
	if _, err := planUp_5_0(tmp); err == nil || !strings.Contains(err.Error(), "mystery") {
		t.Fatalf("expected an error naming mystery/, got %v", err)
	}
	mustExist(t, tmp, ".charter/guides/idioms/go/stdlib-first.md") // nothing moved
}

func TestV5_Down_RestoresCharterLayout(t *testing.T) {
	tmp := seed40Install(t)
	runV4(t, tmp, planUp_5_0)
	runV4(t, tmp, planDown_5_0)

	for _, gone := range []string{".agents", ".keystone"} {
		if dirExists(filepath.Join(tmp, gone)) {
			t.Errorf("%s/ should be gone after Down", gone)
		}
	}
	mustExist(t, tmp,
		".charter/guides/idioms/go/stdlib-first.md",
		".charter/guides/README.md", // rules/ returns as guides/
		".charter/corpus/state/CODEBASE_STATE.md",
		".charter/sensors/build.md",
		".charter/skills/task/SKILL.md", // folded playbook stays a skill
		".charter/adapters/codex/activation.md",
		".charter/policies/acme/guides/x.md",
		".charter/lockfile.json",
		".charter/INDEX.json",
	)
}

// TestV5_StaleScaffold_ArchivedOnlyWhenUnedited — the shipped 4.0.0
// keystone-new-guide moves to the archive and Down brings it back; an
// edited keystone-new-sensor stays where it is.
func TestV5_StaleScaffold_ArchivedOnlyWhenUnedited(t *testing.T) {
	shipped, err := os.ReadFile("testdata/v4.0.0-keystone-new-guide.SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	tmp := seed40Install(t)
	writeAll(t, tmp, map[string]string{
		".charter/skills/keystone-new-guide/SKILL.md":  string(shipped),
		".charter/skills/keystone-new-sensor/SKILL.md": "---\nname: keystone-new-sensor\ndescription: mine\n---\nedited\n",
	})
	runV4(t, tmp, planUp_5_0)
	mustExist(t, tmp, staleArchive_5_0+"/keystone-new-guide/SKILL.md", ".agents/skills/keystone-new-sensor/SKILL.md")
	if dirExists(filepath.Join(tmp, ".agents/skills/keystone-new-guide")) {
		t.Error("the unedited scaffold skill must leave .agents/skills/")
	}
	runV4(t, tmp, planDown_5_0)
	got, _ := os.ReadFile(filepath.Join(tmp, ".charter/skills/keystone-new-guide/SKILL.md"))
	if string(got) != string(shipped) {
		t.Error("Down must return the archived skill unchanged")
	}
	mustExist(t, tmp, ".charter/skills/keystone-new-sensor/SKILL.md")
}
