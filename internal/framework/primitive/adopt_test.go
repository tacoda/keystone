package primitive

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func seedAdoptable(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeAdopt(t, dir, "keystone.json", `{"version":"2","policies":[],"adapters":[],"charter":{"roots":["docs/**/*.md","GLOSSARY.md"]}}`)
	writeAdopt(t, dir, "docs/adr/0001-naming.md", "# Naming\n\nUse kebab-case for ids.\n")
	writeAdopt(t, dir, "docs/adr/0002-deep-modules.md", "# Deep modules\n\nHide state behind small surfaces.\n")
	writeAdopt(t, dir, "GLOSSARY.md", "# Glossary\n\nCharter: the codified context.\n")
	writeAdopt(t, dir, "notes.txt", "not a document\n")
	writeAdopt(t, dir, ".agents/rules/keep.md", "---\nkind: rule\nid: keep\ndescription: x\n---\n")
	return dir
}

func writeAdopt(t *testing.T, dir, rel, body string) {
	t.Helper()
	abs := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestAdopt_IndexesInPlace(t *testing.T) {
	dir := seedAdoptable(t)
	prims, _, err := Walk(dir, ".agents")
	if err != nil {
		t.Fatal(err)
	}
	var adopted []Primitive
	for _, p := range prims {
		if p.Provenance == "repo" {
			adopted = append(adopted, p)
		}
	}
	if len(adopted) != 3 {
		t.Fatalf("adopted %d entries, want 3: %+v", len(adopted), adopted)
	}
	if adopted[0].Kind != "corpus" || adopted[0].ID != "docs/adr/0001-naming" || adopted[0].Path != "docs/adr/0001-naming.md" {
		t.Errorf("first adopted = %+v", adopted[0])
	}
	// Description derives from the first heading.
	if adopted[0].Description != "Naming" {
		t.Errorf("description = %q, want the document title", adopted[0].Description)
	}
}

func TestAdopt_GuardsAndGlobs(t *testing.T) {
	dir := seedAdoptable(t)
	// A doc under .agents/ must never adopt, whatever the roots say.
	writeAdopt(t, dir, ".agents/docs/leak.md", "# Leak\n")
	cfg := `{"version":"2","policies":[],"adapters":[],"charter":{"roots":["**/*.md","**/*.txt"]}}`
	writeAdopt(t, dir, "keystone.json", cfg)
	prims, warns, err := Walk(dir, ".agents")
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range prims {
		if p.Provenance == "repo" && (strings.HasPrefix(p.Path, ".agents/") || p.Path == "keystone.json") {
			t.Errorf("adopted a guarded file: %+v", p)
		}
	}
	// The txt root only matched the txt file: globs filter, not widen
	// into guard dirs.
	for _, p := range prims {
		if p.Provenance == "repo" && p.Path != "notes.txt" && !strings.HasSuffix(p.Path, ".md") {
			t.Errorf("unexpected adoption from the ** roots: %+v", p)
		}
	}
	_ = warns
}

func TestAdopt_NoRootsIsNoop(t *testing.T) {
	dir := t.TempDir()
	writeAdopt(t, dir, "keystone.json", `{"version":"2","policies":[],"adapters":[]}`)
	writeAdopt(t, dir, "docs/x.md", "# X\n")
	prims, _, err := Walk(dir, ".agents")
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range prims {
		if p.Provenance == "repo" {
			t.Errorf("adopted without charter.roots: %+v", p)
		}
	}
}

func TestAdopt_FilesUnmodified(t *testing.T) {
	dir := seedAdoptable(t)
	before, _ := os.ReadFile(filepath.Join(dir, "docs/adr/0001-naming.md"))
	if _, _, err := Walk(dir, ".agents"); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(filepath.Join(dir, "docs/adr/0001-naming.md"))
	if string(before) != string(after) {
		t.Error("adoption modified an adopted file")
	}
}
