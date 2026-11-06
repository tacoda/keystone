package projection

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

var sampleRules = Rules{
	Allow: []string{"Bash(go test:*)", "Bash(make)", "Bash(ls:*)", "Read(./docs/**)", "Read(//etc/**)"},
	Ask:   []string{"Bash(git push:*)"},
	Deny:  []string{"Bash(rm:*)", "Read(./.env*)", "WebFetch(domain:evil.test)", "Bash(curl -X POST:*)", "Bash(curl *)", "Bash(sudo)"},
}

func TestCursorPermissions_TranslatesOnlyExactly(t *testing.T) {
	o, unsupported := cursorPermissions(sampleRules, nil)
	next, err := o.MergeFn(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"Shell(ls)"`, `"Read(./docs/**)"`, `"Shell(rm)"`, `"Shell(sudo)"`, `"Read(./.env*)"`, `"WebFetch(evil.test)"`} {
		if !strings.Contains(string(next), want) {
			t.Errorf("cli.json missing %s:\n%s", want, next)
		}
	}
	// `go test` is two words: Shell(go) would allow all of go — never broaden an allow.
	// Shell(make) would allow make with any arguments — never broaden an allow.
	for _, bad := range []string{"Shell(go)", "Shell(curl)", "Shell(make)", "//etc"} {
		if strings.Contains(string(next), bad) {
			t.Errorf("%s must not translate:\n%s", bad, next)
		}
	}
	for _, want := range []string{"allow: Bash(go test:*)", "allow: Bash(make)", "allow: Read(//etc/**)", "ask: Bash(git push:*)", "deny: Bash(curl -X POST:*)", "deny: Bash(curl *)"} {
		if !contains(unsupported, want) {
			t.Errorf("unsupported missing %q: %v", want, unsupported)
		}
	}
}

func TestCodexPermissions_PrefixRules(t *testing.T) {
	o, unsupported := codexPermissions(sampleRules, nil)
	got := string(o.Content)
	for _, want := range []string{
		`prefix_rule(pattern = ["go", "test"], decision = "allow")`,
		`prefix_rule(pattern = ["git", "push"], decision = "prompt")`,
		`prefix_rule(pattern = ["rm"], decision = "forbidden")`,
		`prefix_rule(pattern = ["curl", "-X", "POST"], decision = "forbidden")`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("rules missing %s:\n%s", want, got)
		}
	}
	// An exact-command allow would broaden into a prefix rule.
	if strings.Contains(got, `["make"]`) {
		t.Errorf("exact allow must not become a prefix rule:\n%s", got)
	}
	// A mid-spec wildcard would become a literal "*" argument and match nothing.
	if strings.Contains(got, `"*"`) {
		t.Errorf("wildcard must not become a literal token:\n%s", got)
	}
	for _, want := range []string{"allow: Bash(make)", "deny: Read(./.env*)", "deny: WebFetch(domain:evil.test)", "deny: Bash(curl *)"} {
		if !contains(unsupported, want) {
			t.Errorf("unsupported missing %q: %v", want, unsupported)
		}
	}
}

// TestMergeBuckets_OwnershipRecord — user entries stay, a dropped keystone
// entry goes, and a user's entry identical to a keystone one is the user's.
func TestMergeBuckets_OwnershipRecord(t *testing.T) {
	prev := []byte(`{"model":"x","permissions":{"deny":["Bash(sudo:*)","Read(./.env*)"]}}`)
	want := map[string][]string{"deny": {"Read(./.env*)", "Bash(rm:*)"}}
	next, err := mergeBuckets(prev, want, nil)
	if err != nil {
		t.Fatal(err)
	}
	record := recordBuckets(prev, want, nil)
	if strings.Join(record, ",") != "deny:Bash(rm:*)" {
		t.Errorf("record = %v; the user's own Read(./.env*) must not be claimed", record)
	}
	// The charter drops both rules: keystone removes only what it added.
	after, _ := mergeBuckets(next, map[string][]string{}, record)
	for _, kept := range []string{"Bash(sudo:*)", "Read(./.env*)", `"model"`} {
		if !strings.Contains(string(after), kept) {
			t.Errorf("user entry %s must stay:\n%s", kept, after)
		}
	}
	if strings.Contains(string(after), "Bash(rm:*)") {
		t.Errorf("keystone's dropped rule must go:\n%s", after)
	}
	if same, _ := mergeBuckets(after, map[string][]string{}, nil); string(same) != string(after) {
		t.Error("a no-op merge must not rewrite the file")
	}
}

// seedCascade writes a project allow and a vendored-policy deny on the same
// command and returns every primitive (project + policy).
func seedCascade(t *testing.T) (string, []primitive.Primitive) {
	t.Helper()
	root := seedCharter(t)
	write(t, root, map[string]string{
		".agents/permissions/project.md":            "---\nid: project\ndescription: x\nallow:\n  - Bash(git push:*)\n---\n",
		".keystone/policies/org/permissions/org.md": "---\nid: org\ndescription: x\ndeny:\n  - Bash(git push:*)\n---\n",
	})
	policy, _, err := primitive.Walk(root, ".keystone/policies/org")
	if err != nil {
		t.Fatal(err)
	}
	return root, append(walk(t, root), policy...)
}

// TestPermissions_ReportedNotDropped — a host without a permission config
// gets a Report naming every rule.
func TestPermissions_ReportedNotDropped(t *testing.T) {
	_, prims := seedCascade(t)
	note := reportNote(Plan(prims, Enabled([]string{"github-copilot"})).Outputs, "github-copilot")
	for _, want := range []string{"deny: Bash(git push:*)", "allow: Bash(git push:*)"} {
		if !strings.Contains(note, want) {
			t.Errorf("Copilot report missing %q: %q", want, note)
		}
	}
}

// TestPermissions_CascadeDenyReachesClaude — the policy deny lands beside
// the project allow; Claude Code evaluates deny first, so deny wins.
func TestPermissions_CascadeDenyReachesClaude(t *testing.T) {
	root, prims := seedCascade(t)
	if _, err := Apply(root, Plan(prims, Enabled(nil)).Outputs, nil, Options{Mode: LinkSymlink}); err != nil {
		t.Fatal(err)
	}
	assertContains(t, root, ".claude/settings.json", `"deny": [`, `"allow": [`, "Bash(git push:*)")
}

// reportNote is host's permission Report note, or "".
func reportNote(outs []Output, host string) string {
	for _, o := range outs {
		if o.Host != host || o.Kind != "permission" {
			continue
		}
		if o.Strategy == Report {
			return o.Note
		}
	}
	return ""
}

func contains(xs []string, x string) bool {
	for _, e := range xs {
		if e == x {
			return true
		}
	}
	return false
}

// TestPermissionBridges_Wired guards the hosts that translate permissions;
// a missing bridge silently turns every rule into a report.
func TestPermissionBridges_Wired(t *testing.T) {
	for _, h := range []Profile{ClaudeCode(), Cursor(), Codex()} {
		if h.Permissions == nil {
			t.Errorf("%s must translate permissions", h.Host)
		}
	}
}

func TestAgentsMD_ReportsUnenforcedPermissions(t *testing.T) {
	root := seedCharter(t)
	write(t, root, map[string]string{".agents/permissions/p.md": "---\nid: p\ndescription: x\ndeny:\n  - Bash(rm:*)\n  - Read(./.env*)\n---\n"})
	md := AgentsMD(walk(t, root), Enabled([]string{"codex", "github-copilot"}))
	for _, want := range []string{
		"**Codex** — loads skills, subagents by itself; open rules, commands from `.agents/` yourself",
		"does **not** enforce 1 permission rule(s)", // Codex: Read(./.env*)
		"does **not** enforce 2 permission rule(s)", // Copilot: both
	} {
		if !strings.Contains(md, want) {
			t.Errorf("AGENTS.md missing %q:\n%s", want, md)
		}
	}
	if strings.Contains(md, "**Claude Code** — loads rules, hooks (inferential), skills, commands, subagents by itself; does **not** enforce") {
		t.Error("Claude Code enforces every rule")
	}
}

// TestPermissions_DisabledHostCleanup — a host dropped from adapters loses
// the entries keystone recorded in its shared file; the user's stay.
func TestPermissions_DisabledHostCleanup(t *testing.T) {
	root := seedCharter(t)
	write(t, root, map[string]string{".cursor/cli.json": `{"permissions":{"deny":["Shell(mine)","Shell(rm)"]}}`})
	managed := map[string][]string{".cursor/cli.json": {"deny:Shell(rm)"}}
	outs := PlanWithState(walk(t, root), Enabled(nil), managed).Outputs
	if _, err := Apply(root, outs, nil, Options{Mode: LinkSymlink}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(root, ".cursor/cli.json"))
	if strings.Contains(string(data), "Shell(rm)") || !strings.Contains(string(data), "Shell(mine)") {
		t.Errorf("want keystone's entry gone and the user's kept:\n%s", data)
	}
}
