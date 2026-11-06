package main

import "testing"

func TestParseToolRun_BothFlagForms(t *testing.T) {
	id, _, args, err := parseToolRun([]string{"grep", "--target", "x y", "--mode=fast"})
	if err != nil {
		t.Fatal(err)
	}
	if got := id + "|" + args["target"] + "|" + args["mode"]; got != "grep|x y|fast" {
		t.Errorf("parsed %s", got)
	}
	if _, _, _, err := parseToolRun([]string{"grep", "--target"}); err == nil {
		t.Error("a flag without a value must fail")
	}
}

func TestToolRun_EndToEnd(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{".agents/tools/hello.md": "---\nid: hello\ndescription: Say hi.\ntransport: cli\nrun: \"test \\\"$KEYSTONE_ARG_NAME\\\" = ada\"\nargs:\n  - name: name\n    required: true\n---\n"})
	if err := runTool([]string{"run", "hello", "--name", "ada", "--dir", root}); err != nil {
		t.Errorf("run: %v", err)
	}
	if err := runTool([]string{"run", "hello", "--name", "bob", "--dir", root}); err == nil {
		t.Error("the handler's non-zero exit must surface")
	}
}
