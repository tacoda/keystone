package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
	"github.com/tacoda/keystone/internal/framework/projection"
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

func TestNewPermission_ScaffoldsCleanRules(t *testing.T) {
	_, path := runNewScaffold(t, "permission", "default", "permissions/default.md")
	parsesAndLintsClean(t, path, "permission")
	data, _ := os.ReadFile(path)
	fm, _, _ := primitive.Parse(string(data))
	if len(fm.Allow) == 0 || len(fm.Deny) == 0 {
		t.Errorf("permission scaffold missing allow/deny lists: %+v", fm)
	}
}

func TestNewHookAndSubagent_ScaffoldClean(t *testing.T) {
	_, hook := runNewScaffold(t, "hook", "go-vet", "hooks/go-vet.md")
	parsesAndLintsClean(t, hook, "hook")
	_, sub := runNewScaffold(t, "subagent", "reviewer", "agents/reviewer.md")
	parsesAndLintsClean(t, sub, "subagent")
}

// TestNewRule_PairsReference — the rule's corpus: entry resolves to the
// reference it scaffolds, so the pair lints with no warning.
func TestNewRule_PairsReference(t *testing.T) {
	root, _ := runNewScaffold(t, "rule", "go/errors", "rules/go/errors.md")
	prims, _, err := primitive.Walk(root, config.DefaultAgentsRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(prims) != 2 {
		t.Fatalf("want rule + reference, got %d primitive(s)", len(prims))
	}
	for _, f := range primitive.Lint(prims) {
		if f.Severity == primitive.FindingError || strings.Contains(f.Message, "does not resolve") {
			t.Errorf("scaffold pair lints: %s", f)
		}
	}
}

func TestNew_LegacyKindsNameTheirReplacement(t *testing.T) {
	for old, now := range newRenamed {
		err := runNew([]string{old, "x", "--dir", t.TempDir()})
		if err == nil || !strings.Contains(err.Error(), "keystone new "+now) {
			t.Errorf("new %s: want an error naming `keystone new %s`, got %v", old, now, err)
		}
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

// TestNewAdapter_WritesLoadableHost — the skeleton is a valid custom host,
// and a built-in host name is refused.
func TestNewAdapter_WritesLoadableHost(t *testing.T) {
	root := t.TempDir()
	if err := runNew([]string{"adapter", "zed", "--dir", root}); err != nil {
		t.Fatal(err)
	}
	hosts, err := projection.LoadCustomHosts(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(hosts) != 1 || hosts[0].Host != "zed" {
		t.Errorf("hosts = %v", hosts)
	}
	if err := runNew([]string{"adapter", "cursor", "--dir", root}); err == nil {
		t.Error("a built-in host name must be refused")
	}
}
