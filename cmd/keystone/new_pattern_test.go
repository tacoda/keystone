package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// runNewScaffold dispatches a `keystone new <verb> <id>` into a temp dir and
// returns the scaffolded file's absolute path.
func runNewScaffold(t *testing.T, verb, id, relPath string) (string, string) {
	t.Helper()
	root := t.TempDir()
	if err := runNew([]string{verb, id, "--dir", root}); err != nil {
		t.Fatalf("keystone new %s %s: %v", verb, id, err)
	}
	return root, filepath.Join(root, config.DefaultHarnessRoot, relPath)
}

// parsesAndLintsClean asserts the scaffolded file parses with the expected
// kind and produces no lint errors.
func parsesAndLintsClean(t *testing.T, path, wantKind string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("scaffold not written at %s: %v", path, err)
	}
	fm, ok, err := primitive.Parse(string(data))
	if err != nil || !ok {
		t.Fatalf("parse %s: ok=%v err=%v", path, ok, err)
	}
	if fm.Kind != wantKind {
		t.Errorf("kind = %q, want %q", fm.Kind, wantKind)
	}
	prim := primitive.Primitive{Frontmatter: fm, Path: path}
	for _, f := range primitive.Lint([]primitive.Primitive{prim}) {
		if f.Severity == primitive.FindingError {
			t.Errorf("scaffold lints with error: %s", f)
		}
	}
}

func TestNewPattern_ScaffoldsCleanProse(t *testing.T) {
	_, path := runNewScaffold(t, "pattern", "tutorial", "patterns/tutorial.md")
	parsesAndLintsClean(t, path, "pattern")
}

func TestNewPosture_ScaffoldsCleanPermissions(t *testing.T) {
	_, path := runNewScaffold(t, "posture", "default", "posture/default.md")
	parsesAndLintsClean(t, path, "posture")
	data, _ := os.ReadFile(path)
	fm, _, _ := primitive.Parse(string(data))
	if len(fm.Allow) == 0 || len(fm.Deny) == 0 {
		t.Errorf("posture scaffold missing allow/deny lists: %+v", fm)
	}
}

func TestNewTool_ScaffoldsCleanCallable(t *testing.T) {
	_, path := runNewScaffold(t, "tool", "grep-symbols", "tools/grep-symbols.md")
	parsesAndLintsClean(t, path, "tool")
	data, _ := os.ReadFile(path)
	fm, _, _ := primitive.Parse(string(data))
	if fm.Transport == "" || fm.Run == "" {
		t.Errorf("tool scaffold missing transport/run: %+v", fm)
	}
}
