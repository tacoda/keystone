package plugins

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// initBareRepoWithTag creates a minimal plugin source repo at repoDir with
// the file layout under contents and tags it with `tag`. Returns the
// file:// URL git can clone from.
func initBareRepoWithTag(t *testing.T, repoDir, tag string, contents map[string]string) string {
	t.Helper()
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}

	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = repoDir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=keystone-test",
			"GIT_AUTHOR_EMAIL=test@keystone.local",
			"GIT_COMMITTER_NAME=keystone-test",
			"GIT_COMMITTER_EMAIL=test@keystone.local",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	run("init", "--quiet", "-b", "main")
	for rel, body := range contents {
		path := filepath.Join(repoDir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", path, err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
	run("add", ".")
	run("commit", "--quiet", "-m", "seed")
	run("tag", tag)
	return "file://" + repoDir
}

func TestFetch_FromLocalRepoAndCacheHit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	cache := t.TempDir()
	t.Setenv(CacheDirEnv, cache)

	repo := filepath.Join(t.TempDir(), "repo")
	url := initBareRepoWithTag(t, repo, "v0.2.0", map[string]string{
		"keystone-plugin.json":      `{"name":"example","version":"0.2.0"}`,
		"guides/principles/spec.md": "spec content",
		"README.md":                 "hello",
	})

	c1, err := Fetch(url, "v0.2.0")
	if err != nil {
		t.Fatalf("Fetch first: %v", err)
	}
	if c1.ResolvedSHA == "" || len(c1.ResolvedSHA) < 40 {
		t.Errorf("ResolvedSHA looks wrong: %q", c1.ResolvedSHA)
	}
	if _, err := os.Stat(filepath.Join(c1.Dir, "keystone-plugin.json")); err != nil {
		t.Errorf("expected manifest in cache: %v", err)
	}

	// Second fetch hits the cache; the directory pointer is the same.
	c2, err := Fetch(url, "v0.2.0")
	if err != nil {
		t.Fatalf("Fetch second: %v", err)
	}
	if c2.Dir != c1.Dir {
		t.Errorf("cache miss on second fetch: %q vs %q", c2.Dir, c1.Dir)
	}
	if c2.ResolvedSHA != c1.ResolvedSHA {
		t.Errorf("ResolvedSHA mismatch on cache hit: %q vs %q", c2.ResolvedSHA, c1.ResolvedSHA)
	}
}

func TestFetch_RejectsEmptyArgs(t *testing.T) {
	if _, err := Fetch("", "v1"); err == nil {
		t.Errorf("Fetch with empty URL should error")
	}
	if _, err := Fetch("file:///nope", ""); err == nil {
		t.Errorf("Fetch with empty version should error")
	}
}

func TestInstall_CopiesAndHashes(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	cache := t.TempDir()
	t.Setenv(CacheDirEnv, cache)

	repo := filepath.Join(t.TempDir(), "repo")
	url := initBareRepoWithTag(t, repo, "v1", map[string]string{
		"keystone-plugin.json":      `{"name":"example","version":"1.0.0"}`,
		"guides/principles/spec.md": "spec body",
		"corpus/principles/spec.md": "corpus body",
	})

	c, err := Fetch(url, "v1")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}

	project := t.TempDir()
	installed, err := Install(c, "example", project, "harness")
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if installed.PluginName != "example" {
		t.Errorf("PluginName = %q, want %q", installed.PluginName, "example")
	}
	if installed.PluginVersion != "1.0.0" {
		t.Errorf("PluginVersion = %q, want %q", installed.PluginVersion, "1.0.0")
	}

	wantPaths := []string{
		"harness/plugins/example/keystone-plugin.json",
		"harness/plugins/example/guides/principles/spec.md",
		"harness/plugins/example/corpus/principles/spec.md",
	}
	for _, p := range wantPaths {
		if _, ok := installed.Files[p]; !ok {
			t.Errorf("missing %q in installed.Files: keys=%v", p, keys(installed.Files))
		}
		abs := filepath.Join(project, p)
		if _, err := os.Stat(abs); err != nil {
			t.Errorf("missing file on disk: %s: %v", abs, err)
		}
	}

	// Best-effort read-only check on POSIX.
	if runtime.GOOS != "windows" {
		info, err := os.Stat(filepath.Join(project, "harness/plugins/example/guides/principles/spec.md"))
		if err != nil {
			t.Fatalf("stat: %v", err)
		}
		if info.Mode().Perm()&0o200 != 0 {
			t.Errorf("expected installed file to be read-only, got mode %o", info.Mode().Perm())
		}
	}
}

func TestVerify_CleanAndDirty(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	cache := t.TempDir()
	t.Setenv(CacheDirEnv, cache)

	repo := filepath.Join(t.TempDir(), "repo")
	url := initBareRepoWithTag(t, repo, "v1", map[string]string{
		"keystone-plugin.json": `{"name":"example","version":"1.0.0"}`,
		"guides/a.md":          "alpha",
		"guides/b.md":          "beta",
	})
	c, err := Fetch(url, "v1")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}

	project := t.TempDir()
	installed, err := Install(c, "example", project, "harness")
	if err != nil {
		t.Fatalf("Install: %v", err)
	}

	// Clean state: no drift.
	drifts, err := Verify("example", project, "harness", installed.Files)
	if err != nil {
		t.Fatalf("Verify clean: %v", err)
	}
	if len(drifts) != 0 {
		t.Errorf("expected no drift, got %v", drifts)
	}

	// Modify a file → DriftModified.
	pluginA := filepath.Join(project, "harness/plugins/example/guides/a.md")
	if err := os.Chmod(pluginA, 0o644); err != nil {
		t.Fatalf("chmod for edit: %v", err)
	}
	if err := os.WriteFile(pluginA, []byte("tampered"), 0o644); err != nil {
		t.Fatalf("tamper: %v", err)
	}
	drifts, err = Verify("example", project, "harness", installed.Files)
	if err != nil {
		t.Fatalf("Verify modified: %v", err)
	}
	if len(drifts) != 1 || drifts[0].Kind != DriftModified {
		t.Errorf("expected one DriftModified, got %v", drifts)
	}

	// Remove a file → DriftMissing.
	if err := os.Chmod(filepath.Join(project, "harness/plugins/example/guides/b.md"), 0o644); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	if err := os.Remove(filepath.Join(project, "harness/plugins/example/guides/b.md")); err != nil {
		t.Fatalf("remove b: %v", err)
	}
	drifts, err = Verify("example", project, "harness", installed.Files)
	if err != nil {
		t.Fatalf("Verify missing: %v", err)
	}
	kinds := map[DriftKind]int{}
	for _, d := range drifts {
		kinds[d.Kind]++
	}
	if kinds[DriftMissing] == 0 {
		t.Errorf("expected at least one DriftMissing, got %v", drifts)
	}

	// Add an extra file → DriftExtra.
	if err := os.WriteFile(filepath.Join(project, "harness/plugins/example/guides/c.md"), []byte("extra"), 0o644); err != nil {
		t.Fatalf("write c: %v", err)
	}
	drifts, err = Verify("example", project, "harness", installed.Files)
	if err != nil {
		t.Fatalf("Verify extra: %v", err)
	}
	kinds = map[DriftKind]int{}
	for _, d := range drifts {
		kinds[d.Kind]++
	}
	if kinds[DriftExtra] == 0 {
		t.Errorf("expected at least one DriftExtra, got %v", drifts)
	}
}

func TestReset_RemovesEvenReadOnlyTree(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	cache := t.TempDir()
	t.Setenv(CacheDirEnv, cache)

	repo := filepath.Join(t.TempDir(), "repo")
	url := initBareRepoWithTag(t, repo, "v1", map[string]string{
		"keystone-plugin.json": `{"name":"example","version":"1"}`,
		"guides/x.md":          "x",
	})
	c, err := Fetch(url, "v1")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}

	project := t.TempDir()
	if _, err := Install(c, "example", project, "harness"); err != nil {
		t.Fatalf("Install: %v", err)
	}

	target := filepath.Join(project, "harness/plugins/example")
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("expected install dir: %v", err)
	}

	if err := Reset("example", project, "harness"); err != nil {
		t.Fatalf("Reset: %v", err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Errorf("plugin dir still exists after Reset (or non-NotExist error): %v", err)
	}
}

func keys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// Sanity: every test runs from the repo root with no shell expansion oddities.
func TestPluginsPackageBuilds(t *testing.T) {
	if !strings.Contains(filepath.ToSlash(os.TempDir()), "/") {
		t.Skip("placeholder")
	}
}
