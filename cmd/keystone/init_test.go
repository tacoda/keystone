package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestInit_FreshScaffoldGoldenFiles is a golden-file smoke test for
// `keystone init` end-to-end. Confirms that a fresh init writes the
// expected set of files into a temp dir: harness layout, project config,
// gitignore, and the agent menu file.
//
// Builds the binary once per run via `go build` (cheap on a warm build
// cache), executes it against an empty tempdir, then walks the result.
// Intentionally asserts *file presence*, not byte-equal content, so
// template improvements during 1.x don't break the regression suite —
// the template-drift check in `keystone doctor` is the right tool for
// content comparisons.
func TestInit_FreshScaffoldGoldenFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e init test in -short mode")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	binDir := t.TempDir()
	bin := filepath.Join(binDir, "keystone")

	// Build from the current cmd/keystone source tree. We're already inside
	// it (this test file) so the package path is "." relative to the cwd
	// the `go test` runner sets to the package directory.
	build := exec.Command("go", "build", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}

	projectDir := t.TempDir()
	cmd := exec.Command(bin, "init", projectDir, "--agent", "codex")
	// stdin is closed by default for exec.Cmd; non-TTY path will be taken,
	// which is what we want for a non-interactive test.
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("keystone init failed: %v\n%s", err, out)
	}

	// Files we expect to exist after a fresh init with --agent codex.
	// The list is the contract: anything important to the install ought
	// to show up here. Per-template *content* drift is not asserted —
	// keystone doctor's template-drift mode is the tool for that.
	wantFiles := []string{
		"keystone.json",
		".gitignore",
		"harness/keystone.lock.json",
		"harness/README.md",
		"harness/guides/README.md",
		"harness/guides/process/spec.md",
		"harness/guides/process/release.md",
		"harness/corpus/README.md",
		"harness/sensors/build.md",
		"harness/sensors/lint.md",
		"harness/sensors/test.md",
		"harness/actions/spec.md",
		"harness/actions/orient.md",
		"harness/actions/verify.md",
		"harness/actions/review.md",
		"harness/playbooks/task.md",
		"harness/adapters/codex/activation.md",
		"harness/adapters/codex/lifecycle.md",
		"harness/adapters/codex/sensors.md",
		"harness/learning/README.md",
		"harness/archive/README.md",
		"harness/corpus/state/INSTALL_PROFILE.md",
		"AGENTS.md", // codex's menu file at the project root
	}

	for _, rel := range wantFiles {
		path := filepath.Join(projectDir, rel)
		if info, err := os.Stat(path); err != nil {
			t.Errorf("missing expected file: %s (%v)", rel, err)
		} else if info.IsDir() {
			t.Errorf("expected file, got directory: %s", rel)
		}
	}

	// Universal-principles must NOT scaffold by default — it's opt-in
	// via --starter universal-principles.
	if _, err := os.Stat(filepath.Join(projectDir, "harness/guides/principles/tdd.md")); !os.IsNotExist(err) {
		t.Errorf("guides/principles/tdd.md should not exist on a default init (got err = %v)", err)
	}

	// Vendored plugins dir must not exist on a fresh install (no plugins
	// declared yet). The .gitignore entry is created regardless so future
	// `keystone install` calls don't accidentally commit plugin content.
	if _, err := os.Stat(filepath.Join(projectDir, "harness/plugins")); !os.IsNotExist(err) {
		t.Errorf("harness/plugins should not exist on a fresh init with no plugins declared (got err = %v)", err)
	}
}
