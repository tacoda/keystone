package projection

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// seedCharter writes a small 5.0 charter and returns the repo root.
func seedCharter(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write(t, root, map[string]string{
		".agents/skills/deploy/SKILL.md":       "---\nname: deploy\ndescription: Ship it.\n---\n# Deploy\n",
		".agents/skills/deploy/scripts/run.sh": "echo run\n",
		".agents/rules/go/style.md":            "---\nid: go/style\ndescription: Go style.\npaths:\n  - \"**/*.go\"\n---\n# Style\n\n## Rules\n- gofmt\n",
		".agents/commands/verify.md":           "---\nid: verify\ndescription: Verify.\n---\n# Verify\n",
		".agents/agents/reviewer.md":           "---\nid: reviewer\ndescription: Reviews.\ntools:\n  - Read\n---\n# Reviewer\n",
		".agents/hooks/build.md":               "---\nid: build\ndescription: Build.\nmode: computational\nevent: Stop\nrun: go build ./...\n---\n",
	})
	return root
}

func write(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, body := range files {
		abs := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func walk(t *testing.T, root string) []primitive.Primitive {
	t.Helper()
	prims, _, err := primitive.Walk(root, ".agents")
	if err != nil {
		t.Fatal(err)
	}
	return prims
}

// plan returns the primitive outputs for hosts; root files (AGENTS.md,
// CLAUDE.md, .aider.conf.yml) are covered by agentsmd_test.go.
func plan(t *testing.T, root string, hosts ...Profile) []Output {
	t.Helper()
	var outs []Output
	for _, o := range Plan(walk(t, root), hosts).Outputs {
		if o.Content == nil {
			outs = append(outs, o)
		}
	}
	return outs
}

// apply runs Apply in symlink mode and fails the test on error.
func apply(t *testing.T, root string, outs []Output, owned []string) ApplyResult {
	t.Helper()
	res, err := Apply(root, outs, owned, Options{Mode: LinkSymlink})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestPlan_ClaudeCode(t *testing.T) {
	got := map[string]Strategy{}
	for _, o := range plan(t, seedCharter(t), ClaudeCode()) {
		got[o.Dest] = o.Strategy
	}
	want := map[string]Strategy{
		".claude/skills/deploy":               Symlink,
		".claude/rules/keystone-go-style.md":  Shim,
		".claude/commands/keystone-verify.md": Shim,
		".claude/agents/keystone-reviewer.md": Shim,
	}
	if len(got) != len(want) {
		t.Errorf("planned %v, want %v (the computational hook has no host file)", got, want)
	}
	for dest, s := range want {
		if got[dest] != s {
			t.Errorf("%s: strategy %q, want %q", dest, got[dest], s)
		}
	}
}

// TestPlan_ReusesLocationsHostsAlreadyRead — a host that reads the
// canonical .agents/skills/ needs no output; one that reads .claude/skills/
// reuses Claude Code's link; reading both draws a duplicate warning.
func TestPlan_ReusesLocationsHostsAlreadyRead(t *testing.T) {
	skillRule := Rule{Strategy: Symlink, Dest: func(p primitive.Primitive) string { return ".other/skills/x" }}
	readsAgents := Profile{Host: "reads-agents", Rules: map[primitive.Kind]Rule{primitive.KindSkill: skillRule},
		Reads: map[primitive.Kind][]string{primitive.KindSkill: {".agents/skills"}}}
	readsClaude := Profile{Host: "reads-claude", Rules: map[primitive.Kind]Rule{primitive.KindSkill: skillRule},
		Reads: map[primitive.Kind][]string{primitive.KindSkill: {".claude/skills", ".agents/skills"}}}

	pr := Plan(walk(t, seedCharter(t)), []Profile{ClaudeCode(), readsAgents, readsClaude})
	outs, warns := pr.Outputs, pr.Warnings
	if len(warns) != 1 || warns[0].Host != "reads-claude" {
		t.Errorf("want one duplicate warning for reads-claude, got %v", warns)
	}
	for _, o := range outs {
		if o.Host != "claude-code" && o.Content == nil {
			t.Errorf("unexpected output for %s: %s", o.Host, o.Dest)
		}
	}
}

func TestApply_SymlinkShimIdempotentAndClean(t *testing.T) {
	root := seedCharter(t)
	outs := plan(t, root, ClaudeCode())
	res := apply(t, root, outs, nil)
	if len(res.Wrote) != 4 {
		t.Errorf("wrote %v, want 4 outputs", res.Wrote)
	}
	link, err := os.Readlink(filepath.Join(root, ".claude/skills/deploy"))
	if err != nil || link != filepath.Join("..", "..", ".agents", "skills", "deploy") {
		t.Errorf("skill link = %q, %v", link, err)
	}
	if _, err := os.Stat(filepath.Join(root, ".claude/skills/deploy/scripts/run.sh")); err != nil {
		t.Errorf("link should expose the whole skill dir: %v", err)
	}
	again := apply(t, root, outs, res.Owned)
	if len(again.Wrote) != 0 || len(again.Unchanged) != 4 {
		t.Errorf("second apply wrote %v, want nothing", again.Wrote)
	}
	if d := Check(root, outs, LinkSymlink); len(d) != 0 {
		t.Errorf("fresh projection drifted: %v", d)
	}
}

func TestApply_RefusesUnownedPath(t *testing.T) {
	root := seedCharter(t)
	write(t, root, map[string]string{".claude/skills/deploy/SKILL.md": "mine\n"})
	if _, err := Apply(root, plan(t, root, ClaudeCode()), nil, Options{Mode: LinkSymlink}); err == nil || !strings.Contains(err.Error(), "did not create") {
		t.Fatalf("expected refusal, got %v", err)
	}
	// Preflight: a refused plan writes nothing at all.
	if _, err := os.Stat(filepath.Join(root, ".claude/commands/keystone-verify.md")); !os.IsNotExist(err) {
		t.Errorf("refused apply must not write other outputs (err=%v)", err)
	}
	if got, _ := os.ReadFile(filepath.Join(root, ".claude/skills/deploy/SKILL.md")); string(got) != "mine\n" {
		t.Error("the user's file must survive")
	}
}

func TestApply_PrunesOnlyOwned(t *testing.T) {
	root := seedCharter(t)
	write(t, root, map[string]string{".claude/skills/mine/SKILL.md": "mine\n"})
	first := apply(t, root, plan(t, root, ClaudeCode()), nil)

	if err := os.Remove(filepath.Join(root, ".agents/commands/verify.md")); err != nil {
		t.Fatal(err)
	}
	second := apply(t, root, plan(t, root, ClaudeCode()), first.Owned)
	if strings.Join(second.Pruned, ",") != ".claude/commands/keystone-verify.md" {
		t.Errorf("pruned %v, want only the removed command's shim", second.Pruned)
	}
	if _, err := os.Stat(filepath.Join(root, ".claude/skills/mine/SKILL.md")); err != nil {
		t.Errorf("user skill must survive prune: %v", err)
	}
}

// TestApply_OverwriteOnlyWhenAsked — an unowned file at a planned path
// (e.g. a 4.x projection never recorded) is refused by default and replaced
// only with --overwrite.
func TestApply_OverwriteOnlyWhenAsked(t *testing.T) {
	root := seedCharter(t)
	write(t, root, map[string]string{".claude/commands/keystone-verify.md": "stale\n"})
	outs := plan(t, root, ClaudeCode())
	if _, err := Apply(root, outs, nil, Options{Mode: LinkSymlink}); err == nil || !strings.Contains(err.Error(), "--overwrite") {
		t.Fatalf("expected refusal naming --overwrite, got %v", err)
	}
	if _, err := Apply(root, outs, nil, Options{Mode: LinkSymlink, Overwrite: true}); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(filepath.Join(root, ".claude/commands/keystone-verify.md")); string(got) == "stale\n" {
		t.Error("--overwrite should replace the unowned file")
	}
}

func TestApply_CopyModeAndDrift(t *testing.T) {
	root := seedCharter(t)
	outs := plan(t, root, ClaudeCode())
	if _, err := Apply(root, outs, nil, Options{Mode: LinkCopy}); err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(filepath.Join(root, ".claude/skills/deploy"))
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Fatalf("copy mode should write a real directory, got mode %v", info.Mode())
	}
	if d := Check(root, outs, LinkCopy); len(d) != 0 {
		t.Errorf("fresh copy drifted: %v", d)
	}
	write(t, root, map[string]string{
		".claude/skills/deploy/SKILL.md":      "edited\n",
		".claude/commands/keystone-verify.md": "edited\n",
	})
	if d := Check(root, outs, LinkCopy); len(d) != 2 {
		t.Errorf("want 2 drifts (edited copy + edited shim), got %v", d)
	}
}

func TestPrunable(t *testing.T) {
	for rel, want := range map[string]bool{
		".claude/skills/x":     true,
		"../outside":           false,
		"/etc/passwd":          false,
		".agents/skills/x":     false,
		".keystone/INDEX.json": false,
	} {
		if prunable(rel) != want {
			t.Errorf("prunable(%q) = %v, want %v", rel, !want, want)
		}
	}
}
