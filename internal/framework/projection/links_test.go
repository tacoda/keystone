package projection

import "testing"

func TestRebaseLinks(t *testing.T) {
	src, dest := ".agents/commands/learn.md", ".cursor/skills/learn/SKILL.md"
	for in, want := range map[string]string{
		"[r](../../.keystone/learning/README.md)": "[r](../../../.keystone/learning/README.md)",
		"[s](../rules/process/spec.md#scope)":     "[s](../../../.agents/rules/process/spec.md#scope)",
		"[u](https://example.com/x.md)":           "[u](https://example.com/x.md)",
		"[a](#local)":                             "[a](#local)",
		"[o](../../../../outside.md)":             "[o](../../../../outside.md)",
		"```\n[c](../x.md)\n```":                  "```\n[c](../x.md)\n```",
	} {
		if got := string(RebaseLinks([]byte(in), src, dest)); got != want {
			t.Errorf("%q → %q, want %q", in, got, want)
		}
	}
}
