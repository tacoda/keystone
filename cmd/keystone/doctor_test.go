package main

import (
	"strings"
	"testing"
)

func TestPrintAdaptersState_OptIn(t *testing.T) {
	root := t.TempDir()
	// No adapters: reported as opt-in, not default.
	writeFiles(t, root, map[string]string{
		"keystone.json": `{"version":"2","policies":[]}`,
	})
	out, err := captureStdout(t, func() error { printAdaptersState(root); return nil })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "no adapters") || !strings.Contains(out, "opt-in") {
		t.Errorf("empty adapters must read as opt-in:\n%s", out)
	}
	// Adapters present: listed as opted in.
	writeFiles(t, root, map[string]string{
		"keystone.json": `{"version":"2","policies":[],"adapters":["claude-code","codex"]}`,
	})
	out, err = captureStdout(t, func() error { printAdaptersState(root); return nil })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "2 adapter(s) opted in") || !strings.Contains(out, "claude-code, codex") {
		t.Errorf("adapters must be reported as opted in:\n%s", out)
	}
}
