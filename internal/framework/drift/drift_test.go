package drift

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	abs := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// The plan's acceptance fixture: one dead link, one vanished path, one
// stale date → exactly three warnings, named by file and reason.
func TestScan_ThreeWarnings(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".agents/rules/process/dead-link.md", `---
kind: rule
id: process/dead-link
description: has a dead link
---
See [the old guide](../guides/gone.md) for details.
`)
	writeFile(t, root, ".agents/rules/process/vanished.md", `---
kind: rule
id: process/vanished
description: governs a gone region
paths:
  - "src/legacy/**"
---
`)
	writeFile(t, root, ".agents/rules/process/stale.md", `---
kind: rule
id: process/stale
description: old verification
verified: 2020-01-01
---
`)
	writeFile(t, root, ".agents/rules/process/clean.md", `---
kind: rule
id: process/clean
description: fine
verified: 2026-11-20
---
[ok](./clean.md)
`)
	writeFile(t, root, ".agents/rules/process/clean.md", `---
kind: rule
id: process/clean2
description: link target
---
`)
	// A real file so "src/legacy/**" has nothing to match and the clean
	// link resolves.
	writeFile(t, root, "cmd/main.go", "package main\n")

	rep, err := Scan(root, Options{Now: time.Date(2026, 11, 27, 0, 0, 0, 0, time.UTC), Days: 90})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if rep.Checked != 4 {
		t.Errorf("checked = %d, want 4", rep.Checked)
	}
	if len(rep.Warnings) != 3 {
		t.Fatalf("warnings = %+v, want exactly 3", rep.Warnings)
	}
	byKind := map[string]Finding{}
	for _, w := range rep.Warnings {
		byKind[w.Kind] = w
	}
	if w, ok := byKind[KindDeadLink]; !ok || w.Path != ".agents/rules/process/dead-link.md" || w.Detail != ".agents/rules/guides/gone.md" {
		t.Errorf("dead-link = %+v", w)
	}
	if w, ok := byKind[KindPathsGone]; !ok || w.Path != ".agents/rules/process/vanished.md" || w.Detail != "src/legacy/**" {
		t.Errorf("paths-gone = %+v", w)
	}
	if w, ok := byKind[KindStale]; !ok || w.Path != ".agents/rules/process/stale.md" || w.Detail != "2020-01-01" {
		t.Errorf("stale = %+v", w)
	}
}

func TestScan_CleanTreeReportsZero(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".agents/rules/process/ok.md", `---
kind: rule
id: process/ok
description: ok
paths:
  - "cmd/**"
verified: 2026-11-27
---
[target](./ok.md)
`)
	writeFile(t, root, "cmd/main.go", "package main\n")
	rep, err := Scan(root, Options{Now: time.Date(2026, 11, 27, 0, 0, 0, 0, time.UTC), Days: 90})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(rep.Warnings) != 0 {
		t.Errorf("warnings = %+v, want zero", rep.Warnings)
	}
}

func TestScan_FileScoping(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".agents/rules/process/a.md", `---
kind: rule
id: process/a
description: a
paths:
  - "cmd/**"
---
[dead](./nope.md)
`)
	writeFile(t, root, ".agents/rules/process/b.md", `---
kind: rule
id: process/b
description: b
paths:
  - "docs/**"
---
[dead](./also-nope.md)
`)
	writeFile(t, root, "cmd/main.go", "package main\n")
	writeFile(t, root, "docs/x.md", "# x\n")

	// Scoped to cmd/**: only rule a is checked.
	rep, err := Scan(root, Options{Files: []string{"cmd/main.go"}, Now: time.Date(2026, 11, 27, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if rep.Checked != 1 || len(rep.Warnings) != 1 || rep.Warnings[0].PrimID != "rule/process/a" {
		t.Errorf("scoped scan = checked %d, warnings %+v", rep.Checked, rep.Warnings)
	}
	// Scoped by naming the primitive's own file: rule b is checked.
	rep, err = Scan(root, Options{Files: []string{".agents/rules/process/b.md"}, Now: time.Date(2026, 11, 27, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if rep.Checked != 1 || len(rep.Warnings) != 1 || rep.Warnings[0].PrimID != "rule/process/b" {
		t.Errorf("self-scoped scan = checked %d, warnings %+v", rep.Checked, rep.Warnings)
	}
}

func TestScan_PolicyVendoredNotChecked(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".keystone/policies/acme/rules/x.md", `---
kind: rule
id: x
description: vendored
verified: 2020-01-01
---
[dead](./gone.md)
`)
	rep, err := Scan(root, Options{Now: time.Date(2026, 11, 27, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if rep.Checked != 0 || len(rep.Warnings) != 0 {
		t.Errorf("policy-vendored files must not be checked: %+v", rep)
	}
}
