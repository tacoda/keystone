package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// seedHook writes a hook primitive under <root>/.harness/hooks/<id>.md.
func seedHook(t *testing.T, root, id, body string) {
	t.Helper()
	path := filepath.Join(root, ".harness", "hooks", id+".md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestHookFire_ComputationalPasses(t *testing.T) {
	root := t.TempDir()
	seedHook(t, root, "ok", "---\nkind: hook\nid: ok\ndescription: x\nmode: computational\nevent: pre-verify\nrun: \"true\"\n---\nbody\n")
	if err := runHookFire([]string{"pre-verify", "--dir", root}); err != nil {
		t.Errorf("expected pass, got %v", err)
	}
}

func TestHookFire_ComputationalNonZeroBlocks(t *testing.T) {
	root := t.TempDir()
	seedHook(t, root, "fail", "---\nkind: hook\nid: fail\ndescription: x\nmode: computational\nevent: pre-verify\nrun: \"exit 1\"\n---\nbody\n")
	if err := runHookFire([]string{"pre-verify", "--dir", root}); err == nil {
		t.Error("expected non-zero hook to block (error), got nil")
	}
}

func TestHookFire_NoMatchIsNoop(t *testing.T) {
	root := t.TempDir()
	seedHook(t, root, "ok", "---\nkind: hook\nid: ok\ndescription: x\nmode: computational\nevent: post-verify\nrun: \"true\"\n---\nbody\n")
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
	seedHook(t, root, "review", "---\nkind: hook\nid: review\ndescription: x\nmode: inferential\nevent: on-review\nagent: reviewer\nreturns: findings\n---\nbody\n")
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
