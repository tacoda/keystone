package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// captureStdout runs fn with os.Stdout redirected and returns what it printed.
func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	runErr := fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	return buf.String(), runErr
}

func TestRunDrift_ReportsWarnings(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		"keystone.json":                      `{"version": "2", "policies": []}`,
		".agents/rules/process/stale.md":     "---\nkind: rule\nid: process/stale\ndescription: old\nverified: 2020-01-01\n---\n",
		".agents/rules/process/dead-link.md": "---\nkind: rule\nid: process/dead-link\ndescription: d\n---\n[go](./gone.md)\n",
		".agents/rules/process/vanished.md":  "---\nkind: rule\nid: process/vanished\ndescription: v\npaths:\n  - \"src/legacy/**\"\n---\n",
		".agents/rules/process/clean.md":     "---\nkind: rule\nid: process/clean\ndescription: c\npaths:\n  - \"cmd/**\"\nverified: 2999-01-01\n---\n",
		"cmd/main.go":                        "package main\n",
	})
	out, err := captureStdout(t, func() error { return runDrift([]string{"--dir", root}) })
	if err != nil {
		t.Fatalf("runDrift: %v", err)
	}
	for _, want := range []string{"3 drift warning(s)", "stale", "dead-link", "paths-gone"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "process/clean") {
		t.Errorf("clean rule must not be named:\n%s", out)
	}
}

func TestRunDrift_CleanTree(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		".agents/rules/process/ok.md": "---\nkind: rule\nid: process/ok\ndescription: ok\npaths:\n  - \"cmd/**\"\n---\n",
		"cmd/main.go":                 "package main\n",
	})
	out, err := captureStdout(t, func() error { return runDrift([]string{"--dir", root}) })
	if err != nil {
		t.Fatalf("runDrift: %v", err)
	}
	if !strings.Contains(out, "no context drift") {
		t.Errorf("output = %q", out)
	}
}

func TestRunDrift_DiffScope(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		".agents/rules/process/scoped.md": "---\nkind: rule\nid: process/scoped\ndescription: s\npaths:\n  - \"cmd/**\"\n---\n[dead](./gone.md)\n",
		".agents/rules/process/other.md":  "---\nkind: rule\nid: process/other\ndescription: o\n---\n[dead](./also-gone.md)\n",
		"cmd/main.go":                     "package main\n",
	})
	for _, args := range [][]string{{"init", "-q"}, {"add", "."}, {"-c", "user.email=t@t", "-c", "user.name=t", "commit", "-qm", "init"}} {
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	// Touch cmd/main.go so --diff scopes to it (rule scoped.md matches cmd/**).
	if err := os.WriteFile(filepath.Join(root, "cmd/main.go"), []byte("package main\n\n// touched\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	files, err := gitDiffFiles(root)
	if err != nil {
		t.Fatalf("gitDiffFiles: %v", err)
	}
	if len(files) != 1 || files[0] != "cmd/main.go" {
		t.Fatalf("gitDiffFiles = %v", files)
	}
	out, err := captureStdout(t, func() error { return runDrift([]string{"--dir", root, "--diff"}) })
	if err != nil {
		t.Fatalf("runDrift --diff: %v", err)
	}
	if !strings.Contains(out, "process/scoped") || strings.Contains(out, "process/other") {
		t.Errorf("--diff scope wrong:\n%s", out)
	}
}

func TestRunDrift_RejectsUnknownFlag(t *testing.T) {
	if err := runDrift([]string{"--bogus"}); err == nil {
		t.Fatal("expected error for unknown flag")
	}
}
