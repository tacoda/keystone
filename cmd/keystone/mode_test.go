package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const modesFixture = `---
kind: rule
id: process/modes
description: 'Active mode: paired'
---
# Pacing Modes

**Active mode: paired**

## paired

Pair programming.
`

func modesProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		".agents/rules/process/modes.md": modesFixture,
	})
	return root
}

func TestMode_ReadCurrent(t *testing.T) {
	root := modesProject(t)
	path, err := modesRulePath(root)
	if err != nil {
		t.Fatalf("modesRulePath: %v", err)
	}
	got, err := currentMode(path)
	if err != nil {
		t.Fatalf("currentMode: %v", err)
	}
	if got != "paired" {
		t.Errorf("mode = %q, want paired", got)
	}
}

func TestMode_SetRewritesLineDescriptionAndIndex(t *testing.T) {
	root := modesProject(t)
	if err := runMode([]string{"solo", "--dir", root}); err != nil {
		t.Fatalf("runMode solo: %v", err)
	}
	body := readRel(t, root, ".agents/rules/process/modes.md")
	if !strings.Contains(body, "**Active mode: solo**") {
		t.Errorf("body line not rewritten:\n%s", body)
	}
	if !strings.Contains(body, "description: 'Active mode: solo'") {
		t.Errorf("frontmatter description not rewritten:\n%s", body)
	}
	if !strings.Contains(body, "## paired\n\nPair programming.") {
		t.Errorf("rest of file must stay untouched:\n%s", body)
	}
	// The index carries the description, so the new mode shows up there.
	lite := readRel(t, root, ".keystone/INDEX.lite.json")
	if !strings.Contains(lite, "Active mode: solo") {
		t.Errorf("INDEX.lite.json does not carry the new mode:\n%s", lite)
	}
	// Reading again reports the new mode.
	path, _ := modesRulePath(root)
	got, err := currentMode(path)
	if err != nil || got != "solo" {
		t.Errorf("currentMode after set = %q, %v; want solo", got, err)
	}
}

func TestMode_RejectsUnknownMode(t *testing.T) {
	root := modesProject(t)
	if err := runMode([]string{"yolo", "--dir", root}); err == nil {
		t.Fatal("expected an error for an unknown mode")
	}
	// Nothing was rewritten.
	if got := readRel(t, root, ".agents/rules/process/modes.md"); !strings.Contains(got, "**Active mode: paired**") {
		t.Errorf("mode line changed on rejected input:\n%s", got)
	}
}

func TestMode_NoModesRuleErrors(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		".agents/rules/process/other.md": "---\nkind: rule\nid: process/other\ndescription: x\n---\n",
	})
	if err := runMode([]string{"--dir", root}); err == nil {
		t.Fatal("expected an error when no process/modes rule exists")
	}
	if _, err := os.Stat(filepath.Join(root, ".keystone")); !os.IsNotExist(err) {
		t.Error("no state dir should be created on the error path")
	}
}
