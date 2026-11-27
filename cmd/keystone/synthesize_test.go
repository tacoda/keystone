package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const acceptedCandidate = `---
captured: 2026-11-27
source: review
status: accepted
proposed-layer: rules/process/no-silent-overwrites
proposed-globs:
  - "cmd/**"
description: 'No silent overwrites of state files'
---

## What happened

An agent overwrote a state file.

## Why it matters

State files are the audit trail.
`

const pendingCandidate = `---
captured: 2026-11-27
source: incident
---

## What happened

Something surprising.
`

func synthesizeFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		".keystone/learning/inbox/2026-11-27T10-00-00-no-silent-overwrites.md": acceptedCandidate,
		".keystone/learning/inbox/2026-11-27T11-00-00-surprise.md":              pendingCandidate,
	})
	return root
}

func TestSynthesize_PromotesAcceptedCandidate(t *testing.T) {
	root := synthesizeFixture(t)
	rep, err := synthesizeProject(root, false)
	if err != nil {
		t.Fatalf("synthesizeProject: %v", err)
	}
	if len(rep.Promoted) != 1 {
		t.Fatalf("promoted = %+v, want 1", rep.Promoted)
	}
	m := rep.Promoted[0]
	if m.Kind != "rule" || m.ID != "process/no-silent-overwrites" {
		t.Errorf("move = %+v", m)
	}
	// The destination primitive exists with the proposed paths: and the
	// candidate body.
	dest := readRel(t, root, ".agents/rules/process/no-silent-overwrites.md")
	for _, want := range []string{"kind: rule", "id: process/no-silent-overwrites", "- cmd/**", "## What happened"} {
		if !strings.Contains(dest, want) {
			t.Errorf("dest missing %q:\n%s", want, dest)
		}
	}
	// The candidate was filed under promoted/, not deleted.
	if _, err := os.Stat(filepath.Join(root, ".keystone/learning/promoted/2026-11-27T10-00-00-no-silent-overwrites.md")); err != nil {
		t.Errorf("promoted record missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".keystone/learning/inbox/2026-11-27T10-00-00-no-silent-overwrites.md")); !os.IsNotExist(err) {
		t.Error("accepted candidate must leave the inbox")
	}
	// The unaccepted candidate stays pending.
	if len(rep.Skipped) != 1 || !strings.Contains(rep.Skipped[0].Reason, "not accepted") {
		t.Errorf("skipped = %+v, want the pending candidate skipped", rep.Skipped)
	}
	if _, err := os.Stat(filepath.Join(root, ".keystone/learning/inbox/2026-11-27T11-00-00-surprise.md")); err != nil {
		t.Errorf("pending candidate must stay in the inbox: %v", err)
	}
	// The index picked the new primitive up.
	idx := readRel(t, root, ".keystone/INDEX.lite.json")
	if !strings.Contains(idx, "no-silent-overwrites") {
		t.Errorf("INDEX.lite.json missing the promoted rule:\n%s", idx)
	}
}

func TestSynthesize_DryRunWritesNothing(t *testing.T) {
	root := synthesizeFixture(t)
	rep, err := synthesizeProject(root, true)
	if err != nil {
		t.Fatalf("synthesizeProject dry-run: %v", err)
	}
	if len(rep.Promoted) != 1 {
		t.Fatalf("dry-run promoted = %+v, want 1 planned", rep.Promoted)
	}
	if _, err := os.Stat(filepath.Join(root, ".agents/rules/process/no-silent-overwrites.md")); !os.IsNotExist(err) {
		t.Error("dry-run must not write the destination")
	}
	if _, err := os.Stat(filepath.Join(root, ".keystone/learning/inbox/2026-11-27T10-00-00-no-silent-overwrites.md")); err != nil {
		t.Error("dry-run must not move the candidate")
	}
	if _, err := os.Stat(filepath.Join(root, ".keystone/INDEX.lite.json")); !os.IsNotExist(err) {
		t.Error("dry-run must not reindex")
	}
}

func TestSynthesize_RefusesExistingDestination(t *testing.T) {
	root := synthesizeFixture(t)
	writeFiles(t, root, map[string]string{
		".agents/rules/process/no-silent-overwrites.md": "---\nkind: rule\nid: process/no-silent-overwrites\ndescription: existing\n---\n",
	})
	rep, err := synthesizeProject(root, false)
	if err != nil {
		t.Fatalf("synthesizeProject: %v", err)
	}
	if len(rep.Promoted) != 0 || len(rep.Skipped) != 2 {
		t.Fatalf("promoted=%+v skipped=%+v", rep.Promoted, rep.Skipped)
	}
	found := false
	for _, s := range rep.Skipped {
		if strings.Contains(s.Reason, "destination exists") {
			found = true
		}
	}
	if !found {
		t.Errorf("want a destination-exists skip, got %+v", rep.Skipped)
	}
	// The existing file is untouched.
	if got := readRel(t, root, ".agents/rules/process/no-silent-overwrites.md"); !strings.Contains(got, "description: existing") {
		t.Errorf("existing destination was overwritten:\n%s", got)
	}
}

func TestSynthesize_HookLayer(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		".keystone/learning/inbox/2026-11-27T12-00-00-commit-gate.md": `---
captured: 2026-11-27
status: accepted
proposed-layer: hook
---

## What happened

A bad commit landed.
`,
	})
	rep, err := synthesizeProject(root, false)
	if err != nil {
		t.Fatalf("synthesizeProject: %v", err)
	}
	if len(rep.Promoted) != 1 || rep.Promoted[0].Kind != "hook" || rep.Promoted[0].ID != "commit-gate" {
		t.Fatalf("promoted = %+v", rep.Promoted)
	}
	dest := readRel(t, root, ".agents/hooks/commit-gate.md")
	if !strings.Contains(dest, "kind: hook") {
		t.Errorf("dest:\n%s", dest)
	}
}
