package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeWorkDoc(t *testing.T, root, rel, gate string) string {
	t.Helper()
	path := filepath.Join(root, workDirRel, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\nkind: document\nid: implementation-plan\ngate: " + gate +
		"\ngates:\n  - draft\n  - in-review\n  - approved\n  - executed\n  - done\n---\n# Plan\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestDocumentPromote_ForwardAdvances(t *testing.T) {
	root := t.TempDir()
	path := writeWorkDoc(t, root, "PORT-1/implementation-plan.md", "draft")

	if err := runDocumentPromote([]string{path, "in-review"}); err != nil {
		t.Fatalf("promote draft→in-review: %v", err)
	}
	got, _ := os.ReadFile(path)
	if !strings.Contains(string(got), "gate: in-review") {
		t.Errorf("gate not advanced; file:\n%s", got)
	}
	// Body + gates list preserved.
	if !strings.Contains(string(got), "# Plan") || !strings.Contains(string(got), "- approved") {
		t.Errorf("promote corrupted the file:\n%s", got)
	}
}

func TestDocumentPromote_BackwardRejected(t *testing.T) {
	root := t.TempDir()
	path := writeWorkDoc(t, root, "PORT-1/implementation-plan.md", "approved")
	err := runDocumentPromote([]string{path, "draft"})
	if err == nil {
		t.Fatal("expected error promoting backward, got nil")
	}
	if !strings.Contains(err.Error(), "forward") {
		t.Errorf("expected forward-transition error, got: %v", err)
	}
	// File unchanged.
	got, _ := os.ReadFile(path)
	if !strings.Contains(string(got), "gate: approved") {
		t.Errorf("backward promote must not write; file:\n%s", got)
	}
}

func TestDocumentPromote_UnknownGateRejected(t *testing.T) {
	root := t.TempDir()
	path := writeWorkDoc(t, root, "PORT-1/implementation-plan.md", "draft")
	err := runDocumentPromote([]string{path, "bogus"})
	if err == nil || !strings.Contains(err.Error(), "not in") {
		t.Fatalf("expected unknown-gate error, got: %v", err)
	}
}

func TestDocumentList_FindsInstances(t *testing.T) {
	root := t.TempDir()
	writeWorkDoc(t, root, "PORT-1/implementation-plan.md", "draft")
	// runDocumentList prints to stdout; just assert it walks without error.
	if err := runDocumentList([]string{root}); err != nil {
		t.Fatalf("list: %v", err)
	}
}
