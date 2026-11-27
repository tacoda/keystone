package mcp

import (
	"reflect"
	"testing"
)

// The MCP tools shell out to the CLI, so the contract is the argv each
// builder produces — assert it exactly.

func TestLearnArgs(t *testing.T) {
	got := learnArgs("/p", "no-silent-overwrites", "an agent overwrote state", "review", "rules/process/x", "cmd/** , docs/**")
	want := []string{"learn", "no-silent-overwrites", "--body", "an agent overwrote state", "--dir", "/p",
		"--source", "review", "--layer", "rules/process/x", "--globs", "cmd/**", "docs/**"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("learnArgs = %v, want %v", got, want)
	}
	// Optional fields omitted → no flags.
	got = learnArgs("/p", "s", "b", "", "", "")
	want = []string{"learn", "s", "--body", "b", "--dir", "/p"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("learnArgs minimal = %v, want %v", got, want)
	}
}

func TestSynthesizeArgs(t *testing.T) {
	got := synthesizeArgs("/p", true)
	want := []string{"synthesize", "--dir", "/p", "--json", "--dry-run"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("synthesizeArgs dry = %v, want %v", got, want)
	}
	got = synthesizeArgs("/p", false)
	want = []string{"synthesize", "--dir", "/p", "--json"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("synthesizeArgs = %v, want %v", got, want)
	}
}

func TestAuditArgs(t *testing.T) {
	got := auditArgs("/p")
	want := []string{"audit", "--dir", "/p", "--json"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("auditArgs = %v, want %v", got, want)
	}
}

func TestModeArgs(t *testing.T) {
	got := modeArgs("/p", "solo")
	want := []string{"mode", "solo", "--dir", "/p", "--json"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("modeArgs set = %v, want %v", got, want)
	}
	got = modeArgs("/p", "")
	want = []string{"mode", "--dir", "/p", "--json"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("modeArgs read = %v, want %v", got, want)
	}
}
