package main

import (
	"io/fs"
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
	cmd := exec.Command(bin, "init", projectDir, "--agent", "codex", "--starter", "workflow")

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

	// 6.0: host projection is opt-in — a fresh init records no adapters
	// (the key is omitted when empty) and nothing outside .agents/,
	// .keystone/, keystone.json, AGENTS.md, .gitignore.
	cfgBytes, err := os.ReadFile(filepath.Join(projectDir, "keystone.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(cfgBytes), "codex") {
		t.Errorf("fresh init must not record adapters:\n%s", cfgBytes)
	}
	if _, err := os.Stat(filepath.Join(projectDir, ".codex")); !os.IsNotExist(err) {
		t.Errorf(".codex/ must not be projected at init (opt-in via `keystone target add`)")
	}
	if _, err := os.Stat(filepath.Join(projectDir, ".claude")); !os.IsNotExist(err) {
		t.Errorf(".claude/ must not be projected at init (opt-in via `keystone target add`)")
	}
	if _, err := os.Stat(filepath.Join(projectDir, "CLAUDE.md")); !os.IsNotExist(err) {
		t.Errorf("CLAUDE.md must not be written at init — only the AGENTS.md block")
	}
	// A fresh install starts on the current layout: no migration is pending.
	assertFileContains(t, projectDir, ".keystone/lockfile.json", `"migrations_applied"`, `"6.0"`)
	assertFileContains(t, projectDir, "AGENTS.md", "<!-- keystone:start -->", "# Charter", "## Query the charter", "keystone_charter_query")

	// Universal-principles and the vendored-policies dir must NOT scaffold
	// by default — both opt-in (`--starter universal-principles`; policies
	// via `keystone new policy`).
	if _, err := os.Stat(filepath.Join(projectDir, ".agents/rules/principles/tdd.md")); !os.IsNotExist(err) {
		t.Errorf("guides/principles/tdd.md should not exist on a default init (got err = %v)", err)
	}
	if _, err := os.Stat(filepath.Join(projectDir, ".agents/policies")); !os.IsNotExist(err) {
		t.Errorf("charter/policies should not exist on a fresh init with no policies declared (got err = %v)", err)
	}
}

// TestInit_WorkflowPackOptIn — decision 15, asserted: a default init
// plants no lifecycle (checked in the golden test); the workflow pack
// does, and legacy implies workflow (the transition kit rides the
// phased lifecycle's documents and rules).
func TestInit_WorkflowPackOptIn(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e init test in -short mode")
	}
	dir, run := initRunner(t)

	workflowInit(t, dir, run, []string{"--agent", "codex", "--starter", "workflow"}, []string{
		".agents/skills/task/SKILL.md", ".agents/skills/keystone-spec/SKILL.md", ".agents/rules/process/spec.md",
	})
}

// workflowInit runs init with initArgs into dir and asserts every
// planted file exists.
func workflowInit(t *testing.T, dir string, run func(...string) (string, error), initArgs, planted []string) {
	t.Helper()
	out, err := run(initArgs...)
	if err != nil {
		t.Fatalf("init %v: %v\n%s", initArgs, err, out)
	}
	for _, rel := range planted {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Errorf("%s: missing after %v: %v", rel, initArgs, err)
		}
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
	scaffold := filepath.Join(dir, ".agents/skills/keystone-new/SKILL.md")
	mine := filepath.Join(dir, ".agents/skills/mine/SKILL.md")
	writeMine(t, scaffold, mine)
	t.Logf("DEBUG after write: mine exists=%v", existsAt(mine))

	mustInit(t, run)
	t.Logf("DEBUG after plain: mine exists=%v scaffold exists=%v", existsAt(mine), existsAt(scaffold))
	if !isMine(scaffold) {
		t.Error("re-init without --overwrite replaced a user-edited file")
	}
	mustInit(t, run, "--overwrite")
	t.Logf("DEBUG after overwrite: mine exists=%v", existsAt(mine))
	if isMine(scaffold) {
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
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
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
// freshInitFiles is the contract for a fresh `init --agent codex`: the
// PLANTED set only. The phased lifecycle (spec, orient, review, verify,
// task, mode, learn, synthesize, audit, debt-review, check-drift,
// allocate, bootstrap) and its rules and agents ride the opt-in
// `--starter workflow` pack; keystone-new / -adapter / -policy and
// keystone-index are scaffold ergonomics and stay planted (decision 15:
// availability over adoption — nothing prescriptive plants by default).
var freshInitFiles = []string{
	"keystone.json",
	".gitignore",
	".keystone/lockfile.json",
	".agents/README.md",
	".agents/rules/README.md",
	".agents/references/README.md",
	".agents/hooks/build.md",
	".agents/hooks/lint.md",
	".agents/hooks/test.md",
	".keystone/learning/README.md",
	".keystone/archive/README.md",
	".agents/skills/README.md",
	".agents/skills/keystone-index/SKILL.md",
	".agents/skills/keystone-new/SKILL.md",
	".agents/skills/keystone-new-adapter/SKILL.md",
	".agents/skills/keystone-new-policy/SKILL.md",
	".agents/agents/README.md",
	".agents/commands/README.md",
	".agents/references/state/INSTALL_PROFILE.md",
	"AGENTS.md", // codex's menu file at the project root
	// Generated by the post-install index + project step. Without
	// these the agent has no way to find the keystone skills.
	".keystone/INDEX.json",
}

func existsAt(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// TestInit_FreshWritesNothingOutsideCore — the 6.0 acceptance check for
// projection demotion: a fresh init touches only .agents/, .keystone/,
// keystone.json, AGENTS.md, .gitignore. Then the opt-in path: `keystone
// target add codex` projects the host and lands a green lockfile.
func TestInit_FreshWritesNothingOutsideCore(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e init test in -short mode")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}
	binDir := t.TempDir()
	bin := filepath.Join(binDir, "keystone")
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	projectDir := t.TempDir()
	if out, err := exec.Command(bin, "init", projectDir, "--agent", "codex", "--starter", "workflow").CombinedOutput(); err != nil {
		t.Fatalf("init: %v\n%s", err, out)
	}
	allowed := map[string]bool{
		".agents": true, ".keystone": true,
		"keystone.json": true, "AGENTS.md": true, ".gitignore": true,
	}
	err := filepath.WalkDir(projectDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == projectDir {
			return nil
		}
		rel, _ := filepath.Rel(projectDir, path)
		top := strings.SplitN(filepath.ToSlash(rel), "/", 2)[0]
		if d.IsDir() {
			if !allowed[top] {
				t.Errorf("fresh init wrote outside the core set: %s", rel)
			}
			return filepath.SkipDir
		}
		if !allowed[filepath.ToSlash(rel)] {
			t.Errorf("fresh init wrote outside the core set: %s", rel)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// Opt-in: target add codex projects .codex/ and records the adapter.
	if out, err := exec.Command(bin, "target", "add", "codex", "--dir", projectDir).CombinedOutput(); err != nil {
		t.Fatalf("target add: %v\n%s", err, out)
	}
	assertFileContains(t, projectDir, "keystone.json", `"codex"`)
	if _, err := os.Stat(filepath.Join(projectDir, ".codex")); err != nil {
		t.Errorf("opt-in target add must project .codex/: %v", err)
	}
	assertFileContains(t, projectDir, ".keystone/lockfile.json", `"codex"`)
}
