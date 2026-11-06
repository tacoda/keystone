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
	return root, filepath.Join(root, config.DefaultAgentsRoot, relPath)
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
	parsesAndLintsClean(t, path, "skill") // 4.x pattern folds into skill
}

func TestNewPosture_ScaffoldsCleanPermissions(t *testing.T) {
	_, path := runNewScaffold(t, "posture", "default", "permissions/default.md")
	parsesAndLintsClean(t, path, "permission") // 4.x posture normalizes
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

// TestNewSkill_ScaffoldsSpecShape — a new skill is Agent Skills spec-valid:
// `name` matches the directory and no keystone-only keys are written.
func TestNewSkill_ScaffoldsSpecShape(t *testing.T) {
	root, _ := runNewScaffold(t, "skill", "deploy-app", "skills/deploy-app/SKILL.md")
	prims, _, err := primitive.Walk(root, config.DefaultAgentsRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(prims) != 1 {
		t.Fatalf("walk found %d primitive(s), want 1", len(prims))
	}
	p := prims[0]
	if got := p.Kind + "/" + p.Name + "/" + p.ID; got != "skill/deploy-app/deploy-app" {
		t.Errorf("kind/name/id = %s, want skill/deploy-app/deploy-app", got)
	}
	if primitive.NeedsSpecFix(p) {
		t.Errorf("scaffold is not in spec shape: keys %v", p.Keys)
	}
	for _, f := range primitive.Lint(prims) {
		if f.Severity == primitive.FindingError {
			t.Errorf("scaffold lints with error: %s", f)
		}
	}
}
