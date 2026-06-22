package aider

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectAider_WritesBothFiles(t *testing.T) {
	root := t.TempDir()
	res, err := ProjectAider(root, "# Agent orientation\n\nbody text\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Wrote) != 2 {
		t.Errorf("expected 2 files written, got %v", res.Wrote)
	}

	conv, err := os.ReadFile(filepath.Join(root, ConventionsRelPath))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(conv), "# Agent orientation") {
		t.Errorf("CONVENTIONS.md missing expected body:\n%s", conv)
	}

	cfg, err := os.ReadFile(filepath.Join(root, ConfigRelPath))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"CONVENTIONS.md",
		"AGENTS.md",
		".keystone/INDEX.lite.json",
	} {
		if !strings.Contains(string(cfg), want) {
			t.Errorf(".aider.conf.yml missing %q\n%s", want, cfg)
		}
	}
}

func TestProjectAider_IsIdempotent(t *testing.T) {
	root := t.TempDir()
	if _, err := ProjectAider(root, "body\n"); err != nil {
		t.Fatal(err)
	}
	res2, err := ProjectAider(root, "body\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(res2.Wrote) != 0 {
		t.Errorf("expected idempotent (no writes), got %v", res2.Wrote)
	}
}
