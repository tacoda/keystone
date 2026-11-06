package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAllocate_OverwriteHandsOver — --overwrite converts, then keystone owns
// the originals of enabled hosts and re-projects without a conflict. Moved
// root-file sections and settings hooks leave their files; everything else
// in them stays, and a disabled host's original is kept.
var overwriteFixture = map[string]string{
	"CLAUDE.md":                  "# P\n\n## Iron laws\n\n- Never push to main.\n\n## Go style\n\nFor `**/*.go`: wrap errors.\n",
	".claude/agents/reviewer.md": "---\nname: reviewer\ndescription: Reviews.\ntools: Read\n---\nReview.\n",
	".cursor/rules/ts.mdc":       "---\nglobs: \"**/*.ts\"\ndescription: TS\n---\nStrict.\n",
	".claude/settings.json": `{"model":"x","hooks":{"Stop":[{"matcher":"","hooks":[{"type":"command","command":"make lint && echo ok"}]}],` +
		`"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":"./guard.sh"}]}]}}`,
}

func TestAllocate_OverwriteHandsOver(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, overwriteFixture)
	if err := runAllocate([]string{"--overwrite", "--dir", root}); err != nil {
		t.Fatalf("allocate --overwrite: %v", err)
	}
	claude := readRel(t, root, "CLAUDE.md")
	if strings.Contains(claude, "## Go style") || !strings.Contains(claude, "## Iron laws") {
		t.Errorf("CLAUDE.md: moved section must leave, iron laws stay:\n%s", claude)
	}
	settings := readRel(t, root, ".claude/settings.json")
	assertHas(t, "settings.json", settings, `"model": "x"`, "./guard.sh", "keystone hook fire Stop", "!make lint")
	if _, err := os.Stat(filepath.Join(root, ".claude/agents/reviewer.md")); !os.IsNotExist(err) {
		t.Error("the Claude subagent original must give way to keystone's projection")
	}
	if !strings.Contains(readRel(t, root, ".claude/agents/keystone-reviewer.md"), "Review.") {
		t.Error("keystone must project the converted subagent")
	}
	if readRel(t, root, ".cursor/rules/ts.mdc") != overwriteFixture[".cursor/rules/ts.mdc"] {
		t.Error("cursor is not enabled: its original must stay untouched")
	}
}

func readRel(t *testing.T, root, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(data)
}

func writeFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, body := range files {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// assertHas checks each want is in text; a want starting with "!" must be
// absent.
func assertHas(t *testing.T, name, text string, wants ...string) {
	t.Helper()
	for _, w := range wants {
		if absent, ok := strings.CutPrefix(w, "!"); ok {
			if strings.Contains(text, absent) {
				t.Errorf("%s must not contain %s:\n%s", name, absent, text)
			}
		} else if !strings.Contains(text, w) {
			t.Errorf("%s missing %s:\n%s", name, w, text)
		}
	}
}
