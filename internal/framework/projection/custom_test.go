package projection

import (
	"strings"
	"testing"
)

// TestCustomHost_ProjectsByItsFile — a host defined in .agents/hosts/
// gets links and copies where its file says, an AGENTS.md line, and gap
// reports for hooks and permissions; it is used only when in adapters.
func TestCustomHost_ProjectsByItsFile(t *testing.T) {
	root := seedCharter(t)
	write(t, root, map[string]string{
		".agents/hosts/zed.json": `{"name":"zed","kinds":{"skill":{"strategy":"link","dir":".zed/skills"},"rule":{"strategy":"copy","dir":".zed/rules"}}}`,
	})
	hosts, err := HostsFor(root, []string{"zed"})
	if err != nil {
		t.Fatal(err)
	}
	if hosts[len(hosts)-1].Host != "zed" {
		t.Fatalf("zed not enabled: %v", hosts)
	}
	prims := walk(t, root)
	if _, err := Apply(root, Plan(prims, hosts).Outputs, nil, Options{Mode: LinkSymlink}); err != nil {
		t.Fatal(err)
	}
	assertContains(t, root, ".zed/rules/go-style.md", "# Style")
	assertContains(t, root, ".zed/skills/deploy/SKILL.md", "deploy")
	md := AgentsMD(prims, hosts)
	if !strings.Contains(md, "- **zed** — loads rules, skills by itself; open commands, subagents from `.agents/` yourself; does **not** run hooks") {
		t.Errorf("AGENTS.md zed line:\n%s", md)
	}
	if off, _ := HostsFor(root, nil); off[len(off)-1].Host == "zed" {
		t.Error("a custom host not in adapters must not be enabled")
	}
}

func TestCustomHost_RejectsUnsafe(t *testing.T) {
	for name, body := range map[string]string{
		"escape":   `{"name":"x","kinds":{"rule":{"strategy":"copy","dir":"../out"}}}`,
		"agents":   `{"name":"x","kinds":{"rule":{"strategy":"copy","dir":".agents/rules"}}}`,
		"strategy": `{"name":"x","kinds":{"rule":{"strategy":"shim","dir":".x"}}}`,
		"kind":     `{"name":"x","kinds":{"hook":{"strategy":"index"}}}`,
		"builtin":  `{"name":"cursor","kinds":{}}`,
		"field":    `{"name":"x","kinds":{},"hooks":true}`,
	} {
		file := "x"
		if name == "builtin" {
			file = "cursor"
		}
		root := seedCharter(t)
		write(t, root, map[string]string{".agents/hosts/" + file + ".json": body})
		if _, err := LoadCustomHosts(root); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}
