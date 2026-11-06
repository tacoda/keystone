package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestInit_FreshScaffoldGoldenFiles is a golden-file smoke test for
// `keystone init` end-to-end. Confirms that a fresh init writes the
// expected set of files into a temp dir: charter layout, project config,
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
	wantFiles := freshInitFiles

	assertFilesExist(t, projectDir, wantFiles)

	// --agent codex selects the codex host: keystone.json adapters, the
	// Codex subagent TOMLs, and the generated AGENTS.md / CLAUDE.md blocks.
	assertFileContains(t, projectDir, "keystone.json", `"codex"`)
	assertFileContains(t, projectDir, ".codex/agents/verifier.toml", "developer_instructions = ")
	assertFileContains(t, projectDir, "AGENTS.md", "<!-- keystone:start -->", "## Constraints", "**Codex**")
	assertFileContains(t, projectDir, "CLAUDE.md", "@AGENTS.md")

	// Universal-principles must NOT scaffold by default — it's opt-in
	// via --starter universal-principles.
	if _, err := os.Stat(filepath.Join(projectDir, ".agents/rules/principles/tdd.md")); !os.IsNotExist(err) {
		t.Errorf("guides/principles/tdd.md should not exist on a default init (got err = %v)", err)
	}

	// Vendored policies dir must not exist on a fresh install (no policies
	// declared yet). The .gitignore entry is created regardless so future
	// `keystone install` calls don't accidentally commit policy content.
	if _, err := os.Stat(filepath.Join(projectDir, ".agents/policies")); !os.IsNotExist(err) {
		t.Errorf("charter/policies should not exist on a fresh init with no policies declared (got err = %v)", err)
	}
}

func assertFilesExist(t *testing.T, root string, rels []string) {
	t.Helper()
	for _, rel := range rels {
		info, err := os.Stat(filepath.Join(root, rel))
		switch {
		case err != nil:
			t.Errorf("missing expected file: %s (%v)", rel, err)
		case info.IsDir():
			t.Errorf("expected file, got directory: %s", rel)
		}
	}
}

func assertFileContains(t *testing.T, root, rel string, wants ...string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Errorf("%s: %v", rel, err)
		return
	}
	for _, w := range wants {
		if !strings.Contains(string(data), w) {
			t.Errorf("%s missing %q", rel, w)
		}
	}
}

// TestAddAdapters_AppendsMissingKeepsOrder — target add / re-init only add
// hosts; existing entries and their order stay, and "generic" is skipped.
func TestAddAdapters_AppendsMissingKeepsOrder(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "keystone.json"), []byte(`{"version":"2","policies":[],"adapters":["cursor"]}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := addAdapters(root, []string{"generic", "codex", "cursor"}); err != nil {
		t.Fatal(err)
	}
	assertFileContains(t, root, "keystone.json", `"cursor",`, `"codex"`)
	data, _ := os.ReadFile(filepath.Join(root, "keystone.json"))
	if strings.Contains(string(data), "generic") {
		t.Errorf("generic must not become an adapter:\n%s", data)
	}
}

// TestInit_OverwriteReplacesInPlace — a re-run keeps user edits; only
// --overwrite replaces keystone's files, and nothing the user added is
// deleted. The 4.x --reset is refused.
func TestInit_OverwriteReplacesInPlace(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e init test in -short mode")
	}
	dir, run := initRunner(t)
	mustInit(t, run)
	spec := filepath.Join(dir, ".agents/rules/process/spec.md")
	mine := filepath.Join(dir, ".agents/rules/mine.md")
	writeMine(t, spec, mine)

	mustInit(t, run)
	if !isMine(spec) {
		t.Error("re-init without --overwrite replaced a user-edited file")
	}
	mustInit(t, run, "--overwrite")
	if isMine(spec) {
		t.Error("--overwrite left keystone's scaffold file unreplaced")
	}
	if !isMine(mine) {
		t.Error("--overwrite touched a file the user added")
	}
	if out, err := run("--reset"); err == nil || !strings.Contains(out, "--overwrite") {
		t.Errorf("--reset must be refused naming --overwrite: %v\n%s", err, out)
	}
}

// initRunner builds the binary and returns a fresh project dir plus a
// function that runs `keystone init <dir> <args>` there.
func initRunner(t *testing.T) (string, func(...string) (string, error)) {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "keystone")
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	dir := t.TempDir()
	return dir, func(args ...string) (string, error) {
		out, err := exec.Command(bin, append([]string{"init", dir}, args...)...).CombinedOutput()
		return string(out), err
	}
}

func mustInit(t *testing.T, run func(...string) (string, error), args ...string) {
	t.Helper()
	if out, err := run(args...); err != nil {
		t.Fatalf("init %v: %v\n%s", args, err, out)
	}
}

func writeMine(t *testing.T, paths ...string) {
	t.Helper()
	for _, p := range paths {
		if err := os.WriteFile(p, []byte("MINE\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func isMine(path string) bool {
	data, err := os.ReadFile(path)
	return err == nil && string(data) == "MINE\n"
}

// freshInitFiles is the contract for a fresh `init --agent codex`: anything
// important to the install shows up here.
var freshInitFiles = []string{
	"keystone.json",
	".gitignore",
	".keystone/lockfile.json",
	".agents/README.md",
	".agents/rules/README.md",
	".agents/rules/process/spec.md",
	".agents/rules/process/release.md",
	".agents/rules/process/runtime-resolution.md",
	".agents/references/process/runtime-resolution.md",
	".agents/references/README.md",
	".agents/hooks/build.md",
	".agents/hooks/lint.md",
	".agents/hooks/test.md",
	".agents/hooks/review-functional.md",
	".agents/commands/spec.md",
	".agents/commands/orient.md",
	".agents/commands/verify.md",
	".agents/commands/review.md",
	".agents/skills/task/SKILL.md",
	".keystone/learning/README.md",
	".keystone/archive/README.md",
	".agents/skills/README.md",
	".agents/skills/keystone-bootstrap/SKILL.md",
	".agents/skills/keystone-spec/SKILL.md",
	".agents/skills/keystone-orient/SKILL.md",
	".agents/skills/keystone-check-drift/SKILL.md",
	".agents/skills/keystone-review/SKILL.md",
	".agents/skills/keystone-debt-review/SKILL.md",
	".agents/skills/keystone-mode/SKILL.md",
	".agents/skills/keystone-learn/SKILL.md",
	".agents/skills/keystone-synthesize/SKILL.md",
	".agents/skills/keystone-audit/SKILL.md",
	".agents/skills/keystone-index/SKILL.md",
	".agents/skills/keystone-verify/SKILL.md",
	".agents/skills/keystone-new/SKILL.md",
	".agents/skills/keystone-allocate/SKILL.md",
	".agents/skills/keystone-new-adapter/SKILL.md",
	".agents/skills/keystone-new-policy/SKILL.md",
	".agents/agents/README.md",
	".agents/commands/README.md",
	".agents/references/state/INSTALL_PROFILE.md",
	"AGENTS.md", // codex's menu file at the project root
	// Generated by the post-install index + project step. Without
	// these the agent has no way to find the keystone skills.
	".keystone/INDEX.json",
	".claude/skills/keystone-bootstrap/SKILL.md",
	".claude/skills/keystone-index/SKILL.md",
	".claude/skills/keystone-verify/SKILL.md",
	".claude/skills/keystone-new/SKILL.md",
}
