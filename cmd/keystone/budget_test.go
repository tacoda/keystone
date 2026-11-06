package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestVerifyBudget_FailsOverCapThenPasses — verify's budget check fails on
// an always-loaded root file over the cap and passes once it is trimmed.
func TestVerifyBudget_FailsOverCapThenPasses(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		"keystone.json": `{"version":"2","budgets":{"always_loaded":{"max_tokens":50}}}`,
		"AGENTS.md":     strings.Repeat("word ", 80),
	})
	if err := checkBudgetCaps(root); err == nil {
		t.Fatal("80 always-loaded tokens against a cap of 50 must fail")
	}
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte(strings.Repeat("word ", 20)), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := checkBudgetCaps(root); err != nil {
		t.Errorf("after trimming: %v", err)
	}
}
