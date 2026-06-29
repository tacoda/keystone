package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestVerify_PreVerifyHookBlocks — a failing pre-verify framework hook blocks
// `keystone verify` (the auto-fire gate).
func TestVerify_PreVerifyHookBlocks(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "keystone.json"), []byte(`{"version":"2","policies":[]}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	seedHook(t, root, "block", "---\nkind: hook\nid: block\ndescription: x\nmode: computational\nevent: pre-verify\nrun: \"exit 1\"\n---\nbody\n")

	if err := verifyWithHooks([]string{"--dir", root}); err == nil {
		t.Error("expected a failing pre-verify hook to block verify")
	}
}

// TestVerify_NoPreVerifyHookProceeds — with no pre-verify hook, the gate is a
// no-op and verify runs to completion (clean tree, no policies).
func TestVerify_NoPreVerifyHookProceeds(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "keystone.json"), []byte(`{"version":"2","policies":[]}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := verifyWithHooks([]string{"--dir", root}); err != nil {
		t.Errorf("expected clean verify with no hooks, got %v", err)
	}
}
