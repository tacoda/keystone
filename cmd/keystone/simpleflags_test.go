package main

import "testing"

func TestParseFlags(t *testing.T) {
	var dir, sensor string
	var fix bool
	vals := map[string]*string{"--dir": &dir, "--sensor": &sensor}
	bools := map[string]*bool{"--fix": &fix}
	if _, err := parseFlags([]string{"--dir", "x", "--sensor=s", "--fix"}, vals, bools, func() {}); err != nil {
		t.Fatal(err)
	}
	if got := dir + "|" + sensor; got != "x|s" {
		t.Errorf("values = %s, want x|s", got)
	}
	if !fix {
		t.Error("--fix not set")
	}
	for _, bad := range [][]string{{"--nope"}, {"pos"}, {"--dir"}} {
		if _, err := parseFlags(bad, vals, bools, func() {}); err == nil {
			t.Errorf("%v: expected an error", bad)
		}
	}
	if help, _ := parseFlags([]string{"-h"}, vals, bools, func() {}); !help {
		t.Error("-h should report help")
	}
}
