package agnostic

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCharterBody_HasCoreSections(t *testing.T) {
	body := CharterBody()
	for _, want := range []string{
		"# Charter",
		"## Read first",
		"## Activate by kind",
		"## Iron laws",
		"## Lifecycle",
		"## Override",
		".charter/INDEX.lite.json",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("CharterBody missing %q", want)
		}
	}
	// CHARTER.md is the canonical source; it must not tell the reader to
	// go read some other file for the orientation.
	if strings.Contains(body, "read CHARTER.md") {
		t.Error("CharterBody should not point at itself")
	}
}

func TestProjectCharterMD_WritesAndIsIdempotent(t *testing.T) {
	root := t.TempDir()
	res, err := ProjectCharterMD(root)
	if err != nil {
		t.Fatalf("ProjectCharterMD: %v", err)
	}
	if !res.Wrote || res.Path != CharterMDRelPath {
		t.Fatalf("first write: wrote=%v path=%q", res.Wrote, res.Path)
	}
	data, err := os.ReadFile(filepath.Join(root, CharterMDRelPath))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(data), "# Charter") {
		t.Errorf("CHARTER.md does not start with the charter heading:\n%s", data)
	}
	res2, err := ProjectCharterMD(root)
	if err != nil {
		t.Fatal(err)
	}
	if res2.Wrote {
		t.Error("second ProjectCharterMD should be a no-op")
	}
}

func TestRenderPointer_ClaudeUsesImport(t *testing.T) {
	claude := RenderPointer(ClaudeCodeProfile())
	for _, want := range []string{"CHARTER.md", "@CHARTER.md", "**Subagents**"} {
		if !strings.Contains(claude, want) {
			t.Errorf("claude pointer missing %q", want)
		}
	}
}

func TestRenderPointer_NonImportHosts(t *testing.T) {
	cursor := RenderPointer(CursorProfile())
	if strings.Contains(cursor, "@CHARTER.md") {
		t.Error("cursor has no import mechanism; pointer must not use @import")
	}
	for _, want := range []string{"**No subagents**", "You **must** read"} {
		if !strings.Contains(cursor, want) {
			t.Errorf("cursor pointer missing %q", want)
		}
	}

	generic := RenderPointer(GenericProfile())
	if !strings.Contains(generic, "## On this host\n") || strings.Contains(generic, "## On this host —") {
		t.Errorf("generic pointer should render a plain 'On this host' heading, got:\n%s", generic)
	}
}
