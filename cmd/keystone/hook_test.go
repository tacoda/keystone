package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// seedHook writes a hook primitive under <root>/.agents/sensors/<id>.md.
func seedHook(t *testing.T, root, id, body string) {
	t.Helper()
	path := filepath.Join(root, ".agents", "hooks", id+".md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestHookFire_ComputationalPasses(t *testing.T) {
	root := t.TempDir()
	seedHook(t, root, "ok", "---\nkind: sensor\nid: ok\ndescription: x\nmode: computational\non: pre-verify\nrun: \"true\"\n---\nbody\n")
	if err := runHookFire([]string{"pre-verify", "--dir", root}); err != nil {
		t.Errorf("expected pass, got %v", err)
	}
}

func TestHookFire_ComputationalNonZeroBlocks(t *testing.T) {
	root := t.TempDir()
	seedHook(t, root, "fail", "---\nkind: sensor\nid: fail\ndescription: x\nmode: computational\non: pre-verify\nrun: \"exit 1\"\n---\nbody\n")
	if err := runHookFire([]string{"pre-verify", "--dir", root}); err == nil {
		t.Error("expected non-zero hook to block (error), got nil")
	}
}

func TestHookFire_NoMatchIsNoop(t *testing.T) {
	root := t.TempDir()
	seedHook(t, root, "ok", "---\nkind: sensor\nid: ok\ndescription: x\nmode: computational\non: post-verify\nrun: \"true\"\n---\nbody\n")
	// Fire a different event — nothing matches, must not error.
	if err := runHookFire([]string{"pre-verify", "--dir", root}); err != nil {
		t.Errorf("no-match fire should be a no-op, got %v", err)
	}
}

func TestHookFire_RequiresEvent(t *testing.T) {
	if err := runHookFire([]string{"--dir", t.TempDir()}); err == nil {
		t.Error("expected error when no event given")
	}
}

func TestHookFire_InferentialEmitsManifest(t *testing.T) {
	root := t.TempDir()
	seedHook(t, root, "review", "---\nkind: sensor\nid: review\ndescription: x\nmode: inferential\non: on-review\nagent: reviewer\nreturns: findings\n---\nbody\n")
	// Inferential dispatch is a manifest, not execution — must not error
	// (keystone cannot run the LLM; the host spawns it).
	if err := runHookFire([]string{"on-review", "--dir", root}); err != nil {
		t.Errorf("inferential fire should not error, got %v", err)
	}
}

func TestFrameworkEventsRegistry(t *testing.T) {
	// Sanity: the dispatcher and lint share this set.
	for _, e := range []string{"pre-verify", "on-gate", "post-command"} {
		if !primitive.IsFrameworkEvent(e) {
			t.Errorf("%q should be a framework event", e)
		}
	}
	if primitive.IsFrameworkEvent("PreToolUse") {
		t.Error("PreToolUse is a host phase, not a framework event")
	}
}

// TestHookFire_HostPhaseFailureExitsTwo — hosts block a tool call only on
// exit 2, so a failing hook on a host phase must carry code 2. A signal
// (keystone-fired) keeps the ordinary error code.
func TestHookFire_HostPhaseFailureExitsTwo(t *testing.T) {
	root := t.TempDir()
	seedHook(t, root, "deny", "---\nkind: hook\nid: deny\ndescription: x\nmode: computational\nevent: PreToolUse\nrun: \"exit 1\"\n---\n")
	seedHook(t, root, "gate", "---\nkind: hook\nid: gate\ndescription: x\nmode: computational\nevent: pre-verify\nrun: \"exit 1\"\n---\n")
	if code := exitCode(runHookFire([]string{"PreToolUse", "--dir", root})); code != 2 {
		t.Errorf("host phase failure exit code = %d, want 2", code)
	}
	if code := exitCode(runHookFire([]string{"pre-verify", "--dir", root})); code != 1 {
		t.Errorf("signal failure exit code = %d, want 1", code)
	}
}

// TestRunComputationalHooks_PassesPayload — the host's event JSON reaches
// each hook on stdin.
func TestRunComputationalHooks_PassesPayload(t *testing.T) {
	hooks := []primitive.Primitive{{Frontmatter: primitive.Frontmatter{ID: "p", Run: `grep -q '"tool_name":"Bash"'`}}}
	if n := runComputationalHooks(t.TempDir(), hooks, hookFireOpts{event: "PreToolUse"}, []byte(`{"tool_name":"Bash"}`)); n != 0 {
		t.Error("hook should have read the payload from stdin")
	}
	if n := runComputationalHooks(t.TempDir(), hooks, hookFireOpts{event: "PreToolUse"}, nil); n != 1 {
		t.Error("without a payload the grep should fail")
	}
}

// TestHookFire_FailsClosed — on a host phase, every error before the hooks
// run exits 2, and the charter is found from a subdirectory.
func TestHookFire_FailsClosed(t *testing.T) {
	root := t.TempDir()
	seedHook(t, root, "deny", "---\nkind: hook\nid: deny\ndescription: x\nmode: computational\nevent: PreToolUse\nrun: \"exit 1\"\n---\n")
	if code := exitCode(runHookFire([]string{"PreToolUse", "--bogus"})); code != 2 {
		t.Errorf("bad flag on a host phase: exit %d, want 2", code)
	}
	if err := runHookFire([]string{"PreToolUse", "--dir", t.TempDir()}); err != nil {
		t.Errorf("explicit --dir with no charter has no hooks: %v", err)
	}
	sub := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(sub)
	if code := exitCode(runHookFire([]string{"PreToolUse"})); code != 2 {
		t.Errorf("hook in the charter above cwd must run and block: exit %d, want 2", code)
	}
	t.Chdir(t.TempDir())
	if code := exitCode(runHookFire([]string{"PreToolUse"})); code != 2 {
		t.Errorf("no charter found on a host phase: exit %d, want 2", code)
	}
}

func TestHookContext_BoundsHostPhases(t *testing.T) {
	ctx, cancel := hookContext(hookFireOpts{event: "PreToolUse"})
	defer cancel()
	if _, ok := ctx.Deadline(); !ok {
		t.Error("host-phase hooks need a deadline below the host's timeout")
	}
	sig, cancel2 := hookContext(hookFireOpts{event: "pre-verify"})
	defer cancel2()
	if _, ok := sig.Deadline(); ok {
		t.Error("keystone-fired signals run unbounded")
	}
}
